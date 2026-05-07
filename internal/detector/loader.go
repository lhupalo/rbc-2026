package detector

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
)

type normConstants struct {
	MaxAmount            float64 `json:"max_amount"`
	MaxInstallments      float64 `json:"max_installments"`
	AmountVsAvgRatio     float64 `json:"amount_vs_avg_ratio"`
	MaxMinutes           float64 `json:"max_minutes"`
	MaxKm                float64 `json:"max_km"`
	MaxTxCount24h        float64 `json:"max_tx_count_24h"`
	MaxMerchantAvgAmount float64 `json:"max_merchant_avg_amount"`
}

type ivfConfig struct {
	NClusters      int     `json:"n_clusters"`
	Nprobe         int     `json:"nprobe"`
	KNeighbors     int     `json:"k_neighbors"`
	FraudThreshold float64 `json:"fraud_threshold"`
	VectorDim      int     `json:"vector_dim"`
}

func Load(dataDir string) (*Detector, error) {
	norm, err := loadNorm(dataDir + "/normalization.json")
	if err != nil {
		return nil, fmt.Errorf("normalization: %w", err)
	}

	mcc, err := loadMCC(dataDir + "/mcc_risk.json")
	if err != nil {
		return nil, fmt.Errorf("mcc_risk: %w", err)
	}

	cfg, err := loadIVFConfig(dataDir + "/ivf_config.json")
	if err != nil {
		return nil, fmt.Errorf("ivf_config: %w", err)
	}

	centroids, nClusters, dim, err := loadCentroids(dataDir + "/centroids.bin")
	if err != nil {
		return nil, fmt.Errorf("centroids: %w", err)
	}

	vectors, labels, err := loadVectors(dataDir+"/vectors.bin", dim)
	if err != nil {
		return nil, fmt.Errorf("vectors: %w", err)
	}

	invLists, err := loadIVFStructure(dataDir+"/ivf_structure.bin", nClusters)
	if err != nil {
		return nil, fmt.Errorf("ivf_structure: %w", err)
	}

	index := newIVFIndex(centroids, vectors, labels, invLists,
		dim, nClusters, cfg.Nprobe, cfg.KNeighbors, cfg.FraudThreshold)

	return &Detector{
		index: index,
		mcc:   mcc,
		norm:  norm,
	}, nil
}

func loadNorm(path string) (normConstants, error) {
	f, err := os.Open(path)
	if err != nil {
		return normConstants{}, err
	}
	defer f.Close()

	var n normConstants
	if err := json.NewDecoder(f).Decode(&n); err != nil {
		return normConstants{}, err
	}
	return n, nil
}

func loadMCC(path string) (map[string]float64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var m map[string]float64
	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return nil, err
	}
	return m, nil
}

func loadIVFConfig(path string) (ivfConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return ivfConfig{}, err
	}
	defer f.Close()

	var cfg ivfConfig
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return ivfConfig{}, err
	}
	return cfg, nil
}

// loadCentroids lê centroids.bin: [int64 n_clusters][int64 dim][float64 × n_clusters × dim]
// Os float64 são quantizados para int8 na leitura, reduzindo ~445KB → ~55KB.
func loadCentroids(path string) (data []int8, nClusters, dim int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 1<<20)

	var nc, d int64
	if err = binary.Read(r, binary.LittleEndian, &nc); err != nil {
		return nil, 0, 0, err
	}
	if err = binary.Read(r, binary.LittleEndian, &d); err != nil {
		return nil, 0, 0, err
	}

	f64s, err := readFloat64s(r, int(nc*d))
	if err != nil {
		return nil, 0, 0, err
	}
	data = make([]int8, len(f64s))
	for i, v := range f64s {
		data[i] = quantizeFloat64(v)
	}
	return data, int(nc), int(d), nil
}

// loadVectors lê vectors.bin em chunks e quantiza cada float64 para int8 imediatamente,
// evitando manter o array float64 completo (~336MB) em RAM.
// Formato: [int64 n_vectors][int64 dim][float64 × n_vectors × dim][uint8 × n_vectors]
func loadVectors(path string, expectedDim int) (vecs []int8, labels []bool, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 1<<20)

	var nVec, dim int64
	if err = binary.Read(r, binary.LittleEndian, &nVec); err != nil {
		return nil, nil, err
	}
	if err = binary.Read(r, binary.LittleEndian, &dim); err != nil {
		return nil, nil, err
	}
	if int(dim) != expectedDim {
		return nil, nil, fmt.Errorf("vectors.bin: dim=%d diverge do esperado %d", dim, expectedDim)
	}

	vecs, err = readQuantizedVectors(r, int(nVec), int(dim))
	if err != nil {
		return nil, nil, err
	}

	rawLabels := make([]uint8, nVec)
	if err = binary.Read(r, binary.LittleEndian, rawLabels); err != nil {
		return nil, nil, err
	}

	labels = make([]bool, nVec)
	for i, l := range rawLabels {
		labels[i] = l == 1
	}
	return vecs, labels, nil
}

// readQuantizedVectors lê nVec vetores de dim float64 em chunks e quantiza para int8.
// Pico de memória por chunk: chunkSize × dim × 8 bytes (float64 temporário).
func readQuantizedVectors(r *bufio.Reader, nVec, dim int) ([]int8, error) {
	const chunkVecs = 4096
	buf := make([]byte, chunkVecs*dim*8)
	out := make([]int8, nVec*dim)

	remaining := nVec
	offset := 0
	for remaining > 0 {
		batch := remaining
		if batch > chunkVecs {
			batch = chunkVecs
		}
		b := buf[:batch*dim*8]
		if _, err := readFull(r, b); err != nil {
			return nil, err
		}
		for j := 0; j < batch*dim; j++ {
			bits := binary.LittleEndian.Uint64(b[j*8:])
			v := math.Float64frombits(bits)
			out[offset+j] = quantizeFloat64(v)
		}
		offset += batch * dim
		remaining -= batch
	}
	return out, nil
}

// loadIVFStructure lê ivf_structure.bin: [int64 n_clusters] depois para cada cluster [int64 list_size][int32 × list_size]
func loadIVFStructure(path string, nClusters int) ([][]int32, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, 1<<20)

	var n int64
	if err = binary.Read(r, binary.LittleEndian, &n); err != nil {
		return nil, err
	}
	if int(n) != nClusters {
		return nil, fmt.Errorf("ivf_structure.bin: n_clusters=%d diverge do esperado %d", n, nClusters)
	}

	invLists := make([][]int32, n)
	for i := int64(0); i < n; i++ {
		var listSize int64
		if err = binary.Read(r, binary.LittleEndian, &listSize); err != nil {
			return nil, fmt.Errorf("cluster %d: %w", i, err)
		}
		ids := make([]int32, listSize)
		if err = binary.Read(r, binary.LittleEndian, ids); err != nil {
			return nil, fmt.Errorf("cluster %d ids: %w", i, err)
		}
		invLists[i] = ids
	}
	return invLists, nil
}

// readFloat64s lê n float64s little-endian de r em chunks para eficiência.
func readFloat64s(r *bufio.Reader, n int) ([]float64, error) {
	const chunkElems = 8192
	buf := make([]byte, chunkElems*8)
	out := make([]float64, n)
	i := 0
	for i < n {
		batch := n - i
		if batch > chunkElems {
			batch = chunkElems
		}
		b := buf[:batch*8]
		if _, err := readFull(r, b); err != nil {
			return nil, err
		}
		for j := 0; j < batch; j++ {
			bits := binary.LittleEndian.Uint64(b[j*8:])
			out[i+j] = math.Float64frombits(bits)
		}
		i += batch
	}
	return out, nil
}

func readFull(r *bufio.Reader, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := r.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
