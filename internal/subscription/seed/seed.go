package seed

import (
	"io/ioutil"
	"nats-server/internal/config"
	"os"
	"path"
	"runtime"

	"github.com/nats-io/stan.go"
)

func Seed(cfg config.NatsConfig) ([][]byte, error) {
	result := [][]byte{}

	_, b, _, _ := runtime.Caller(0)
	p := path.Join(path.Dir(b), "data")

	sc, err := stan.Connect(cfg.ClusterID, "pub", stan.NatsURL(cfg.Port))
	if err != nil {
		return nil, err
	}
	defer sc.Close()

	files, err := ioutil.ReadDir(p)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		data, err := os.ReadFile(path.Join(p, file.Name()))
		if err != nil {
			return nil, err
		}

		if err := sc.Publish("orders", data); err != nil {
			return nil, err
		}

		result = append(result, data)
	}

	return result, nil
}
