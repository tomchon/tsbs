package tdengine

import (
	"bytes"
	"math"

	"github.com/timescale/tsbs/pkg/data/source"
	"github.com/timescale/tsbs/pkg/targets"
)

func NewBenchmark(dbName string, opts *LoadingOptions, dataSourceConfig *source.DataSourceConfig) (targets.Benchmark, error) {
	var ds targets.DataSource
	if dataSourceConfig.Type == source.FileDataSourceType {
		ds = newFileDataSource(dataSourceConfig.File.Location)
	} else {
		panic("not implement")
	}

	return &benchmark{
		opts:   opts,
		ds:     ds,
		dbName: dbName,
	}, nil
}

type benchmark struct {
	opts   *LoadingOptions
	ds     targets.DataSource
	dbName string
}

func (b *benchmark) GetDataSource() targets.DataSource {
	return b.ds
}

func (b *benchmark) GetBatchFactory() targets.BatchFactory {
	return &factory{}
}

func (b *benchmark) GetPointIndexer(maxPartitions uint) targets.PointIndexer {
	if maxPartitions > 1 {
		interval := uint32(math.MaxUint32 / maxPartitions)
		hashEndGroups := make([]uint32, maxPartitions)
		for i := 0; i < int(maxPartitions); i++ {
			if i == int(maxPartitions)-1 {
				hashEndGroups[i] = math.MaxUint32
			} else {
				hashEndGroups[i] = interval*uint32(i+1) - 1
			}
		}
		prefix := []byte("1." + b.dbName + ".")
		return &indexer{buffer: &bytes.Buffer{}, prefix: prefix, hashEndGroups: hashEndGroups, partitions: int(maxPartitions), tmp: map[string]uint{}}
	}
	return &targets.ConstantIndexer{}
}

func (b *benchmark) GetProcessor() targets.Processor {
	return newProcessor(b.opts, b.dbName)
}

func (b *benchmark) GetDBCreator() targets.DBCreator {
	return &dbCreator{opts: b.opts, ds: b.ds}
}
