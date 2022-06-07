package contract

import (
	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/contracts/staking"
)

func NewStakingAccount(
	filepath string,
	constructorParams []interface{},
) (*chain.GenesisAccount, error) {
	//	#1: create artifact from contract json file
	artifact, err := NewArtifact(filepath)
	if err != nil {
		return nil, err
	}

	//	#2: verify required methods are present (staking)
	if err := artifact.Implements(
		"Stake",
		"Unstake",
		"GetValidatorSet",
	); err != nil {
		return nil, err
	}

	//	TODO (milos): where does config come from ?
	config := chain.ForksInTime{
		Homestead:      true,
		Byzantium:      true,
		Constantinople: true,
		Petersburg:     true,
		Istanbul:       true,
		EIP150:         true,
		EIP158:         true,
		EIP155:         true,
	}

	//	#3: create new genesis account based on the artifact and user-provided params
	return NewContractAccount(config, staking.AddrStakingContract, artifact, constructorParams)
}
