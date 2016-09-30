package commands

import (
	"github.com/Shopify/themekit/kit"
)

func drainErrors(errs chan error) {
	for {
		if err := <-errs; err != nil {
			kit.NotifyError(err)
		} else {
			break
		}
	}
}
