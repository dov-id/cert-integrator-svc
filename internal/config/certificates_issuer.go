package config

import (
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type Contract struct {
	Address   string `figure:"address"`
	Name      string `figure:"name"`
	FromBlock int64  `figure:"from_block"`
}

type ContractsCfg struct {
	List []Contract
}

func (c *config) CertificatesIssuer() *ContractsCfg {
	return c.certificatesIssuer.Do(func() interface{} {
		var cfg ContractsCfg

		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, ContractHooks).
			From(kv.MustGetStringMap(c.getter, "certificates_issuer")).
			Please()

		if err != nil {
			panic(errors.Wrap(err, "failed to figure out certificates issuer config"))
		}

		return &cfg
	}).(*ContractsCfg)
}
