package ecsx

import "github.com/telkomindonesia/crs-offline/ecs"

type Document struct {
	ecs.Document
	HTTP *HTTP `json:"http"`

	CRS *CRS `json:"_crs"`
}
