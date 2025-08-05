package adapters

import (
	"github.com/google/uuid"
)

type NodeIdentityRepo struct {
	uuid uuid.UUID
}

func NewNodeIdentityRepo() *NodeIdentityRepo {
	return &NodeIdentityRepo{uuid: uuid.New()}
}

func (r *NodeIdentityRepo) NodeUUID() uuid.UUID {
	return r.uuid
}
