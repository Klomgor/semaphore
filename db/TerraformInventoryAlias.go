package db

import "reflect"

type TerraformInventoryAlias struct {
	ID          int    `db:"id" json:"id"`
	ProjectID   int    `db:"project_id" json:"project_id"`
	InventoryID int    `db:"inventory_id" json:"inventory_id"`
	AuthKeyID   int    `db:"auth_key_id" json:"auth_key_id"`
	Alias       string `db:"alias" json:"alias"`
}

var TerraformInventoryAliasProps = ObjectProps{
	TableName:         "project__terraform_inventory_alias",
	Type:              reflect.TypeOf(TerraformInventoryAlias{}),
	PrimaryColumnName: "id",
}

func (alias TerraformInventoryAlias) ToAlias() Alias {
	return Alias{
		ID:        alias.ID,
		Alias:     alias.Alias,
		ProjectID: alias.ProjectID,
	}
}
