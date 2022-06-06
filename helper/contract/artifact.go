package contract

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/umbracle/ethgo/abi"

	"github.com/0xPolygon/polygon-edge/chain"
	"github.com/0xPolygon/polygon-edge/helper/hex"
	"github.com/0xPolygon/polygon-edge/state"
	itrie "github.com/0xPolygon/polygon-edge/state/immutable-trie"
	"github.com/0xPolygon/polygon-edge/types"
)

type Artifact struct {
	*abi.ABI

	Bytecode,
	DeployedBytecode []byte
}

func NewArtifact(filepath string) (*Artifact, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	artifact := new(Artifact)
	if err := artifact.initialize(bytes); err != nil {
		return nil, err
	}

	return artifact, nil
}

func NewContractAccount(
	forks chain.ForksInTime,
	addr types.Address,
	artifact *Artifact,
	constructorParams ...interface{},
) (*chain.GenesisAccount, error) {
	//	TODO: (milos) do constructors params need to be present ?
	if constructorParams == nil {
		return nil, errors.New("no constructor parameters provided")
	}

	initCode, err := artifact.EncodeConstructor(constructorParams)
	if err != nil {
		return nil, fmt.Errorf("unable to encode params: err=%w", err)
	}

	return state.GenerateContractAccount(
		forks,
		addr,
		initCode,
		itrie.NewState(itrie.NewMemoryStorage()),
	)
}

func (a *Artifact) EncodeConstructor(params ...interface{}) ([]byte, error) {
	constructor, err := abi.Encode(params, a.Constructor.Inputs)
	if err != nil {
		return nil, err
	}

	code := append(a.Bytecode, constructor...)

	return code, nil
}

func (a *Artifact) Implements(methods ...string) error {
	for _, method := range methods {
		if a.ABI.GetMethod(method) == nil {
			return fmt.Errorf("implementation missing: method=%v", method)
		}
	}

	return nil
}

func (a *Artifact) initialize(bytes []byte) error {
	var fileJSON map[string]interface{}
	if err := json.Unmarshal(bytes, &fileJSON); err != nil {
		return err
	}

	/*	parse abi */
	if err := a.setABI(fileJSON); err != nil {
		return err
	}

	/*	parse bytecode */
	if err := a.setBytecode(fileJSON); err != nil {
		return err
	}

	/*	parse deployed bytecode */
	if err := a.setDeployedBytecode(fileJSON); err != nil {
		return err
	}

	return nil
}

func (a *Artifact) setABI(jsonMap map[string]interface{}) error {
	rawABI, ok := jsonMap["contractABI"]
	if !ok {
		panic("bad")
	}

	contractABI, err := json.Marshal(rawABI)
	if err != nil {
		return err
	}

	if a.ABI, err = abi.NewABI(string(contractABI)); err != nil {
		return err
	}

	return nil
}

func (a *Artifact) setBytecode(jsonMap map[string]interface{}) error {
	rawBytecode, ok := jsonMap["bytecode"].(string)
	if !ok {
		panic("bad")
	}

	bytecode, err := hex.DecodeString(strings.TrimPrefix(rawBytecode, "0x"))
	if err != nil {
		return err
	}

	a.Bytecode = bytecode

	return nil
}

func (a *Artifact) setDeployedBytecode(jsonMap map[string]interface{}) error {
	rawDeployedBytecode, ok := jsonMap["deployedBytecode"].(string)
	if !ok {
		panic("bad ")
	}

	deployedBytecode, err := hex.DecodeString(strings.TrimPrefix(rawDeployedBytecode, "0x"))
	if err != nil {
		return err
	}

	a.DeployedBytecode = deployedBytecode

	return nil
}
