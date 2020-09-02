package contracts

import (
	"github.com/enorith/database"
	"github.com/enorith/database/rithythm"
)

type User interface {
	UserIdentifier() uint64
	UserIdentifierName() string
	Unmarshal(data *database.CollectionItem)
	CloneUser() User
}

type RithythmUser interface {
	rithythm.DataModel
	UserIdentifier() uint64
	UserIdentifierName() string
	CloneUser() User
}
