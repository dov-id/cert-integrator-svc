package handlers

import (
	"fmt"
	"net/http"

	"github.com/dov-id/cert-integrator-svc/internal/data"
	"github.com/dov-id/cert-integrator-svc/internal/data/postgres"
	"github.com/dov-id/cert-integrator-svc/internal/service/api/models"
	"github.com/dov-id/cert-integrator-svc/internal/service/api/requests"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-merkletree-sql/v2"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
)

func GenerateSMTProof(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewGenerateProofRequest(r)
	if err != nil {
		Log(r).WithError(err).Error("failed to parse generate proof request")
		ape.RenderErr(w, problems.BadRequest(err)...)
		return
	}

	contract, err := ContractsQ(r).FilterByAddresses(request.Data.Attributes.Contract).Get()
	if err != nil {
		Log(r).WithError(err).Error("failed to get contract by address")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	if contract == nil {
		Log(r).WithError(err).Errorf("no contract with address `%s`", request.Data.Attributes.Contract)
		ape.RenderErr(w, problems.NotFound())
		return
	}

	treeStorage := postgres.NewStorage(DB(r), contract.Id)
	mTree, err := merkletree.NewMerkleTree(r.Context(), treeStorage, data.MaxMTreeLevel)
	if err != nil {
		Log(r).WithError(err).Error("failed to open/create merkle tree")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	keyBig, err := merkletree.NewBigIntFromHashBytes(common.HexToAddress(request.Data.Attributes.NodeKey).Hash().Bytes())
	if err != nil {
		Log(r).WithError(err).Error("failed to convert node key to hash")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	proof, value, err := mTree.GenerateProof(r.Context(), keyBig, mTree.Root())
	if err != nil {
		Log(r).WithError(err).Error("failed to generate merkle tree proof")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	var hexProof = make([]string, 0)
	for _, sibling := range proof.AllSiblings() {
		hexProof = append(hexProof, fmt.Sprintf("0x%s", sibling.Hex()))
	}

	keyHash, err := merkletree.NewHashFromBigInt(keyBig)
	if err != nil {
		Log(r).WithError(err).Error("failed to convert key big int to hash")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	valueHash, err := merkletree.NewHashFromBigInt(value)
	if err != nil {
		Log(r).WithError(err).Error("failed to convert value big int to hash")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	ape.Render(w, models.NewLinkResponse(int64(contract.Id), keyHash.Hex(), valueHash.Hex(), hexProof))
}