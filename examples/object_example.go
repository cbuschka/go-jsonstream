package examples

import (
	"github.com/cbuschka/go-jsonstream"
	"os"
)

func run() error {

	wr := jsonstream.NewWriter(os.Stdout)
	defer func() {
		_ = wr.Close()
	}()
	if err := wr.WriteObjectStart(); err != nil {
		return err
	}

	if err := wr.WriteKeyAndStringValue("key", "value"); err != nil {
		return err
	}

	if err := wr.WriteObjectEnd(); err != nil {
		return err
	}

	return nil
}
