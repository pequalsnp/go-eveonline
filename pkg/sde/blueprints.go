package sde

import (
	"fmt"

	"github.com/pequalsnp/go-eveonline/pkg/eveonline"
	yaml "gopkg.in/yaml.v2"
)

type ProductBlueprintMap map[eveonline.TypeID][]*Blueprint

type TypeQuantity struct {
	TypeID   eveonline.TypeID `yaml:"typeID"`
	Quantity int              `yaml:"quantity"`
}

type SkillLevel struct {
	SkillID eveonline.TypeID `yaml:"typeID"`
	Level   int              `yaml:"level"`
}

func (tq TypeQuantity) String() string {
	return fmt.Sprintf("type %d quantity %d", tq.TypeID, tq.Quantity)
}

type Reaction struct {
	Materials []TypeQuantity `yaml:"materials"`
	Products  []TypeQuantity `yaml:"products"`
}

type Manufacturing struct {
	Materials      []TypeQuantity `yaml:"materials"`
	Products       []TypeQuantity `yaml:"products"`
	TimeInSeconds  int            `yaml:"time"`
	RequiredSkills []SkillLevel   `yaml:"skills"`
}

type Activities struct {
	Reaction      *Reaction      `yaml:"reaction"`
	Manufacturing *Manufacturing `yaml:"manufacturing"`
}

func (a Activities) String() string {
	return fmt.Sprintf("%+v", a.Reaction)
}

type Blueprint struct {
	BlueprintTypeID eveonline.TypeID `yaml:"blueprintTypeID"`
	Activities      *Activities      `yaml:"activities"`
}

func (b Blueprint) ProductsAndInputs() ([]TypeQuantity, []TypeQuantity) {
	if b.IsReaction() {
		return b.Activities.Reaction.Products, b.Activities.Reaction.Materials
	} else {
		return b.Activities.Manufacturing.Products, b.Activities.Manufacturing.Materials
	}
}

func (b *Blueprint) CreatesProducts() ([]eveonline.TypeID, error) {
	products := make([]eveonline.TypeID, 0)
	if b.Activities.Manufacturing != nil {
		for _, product := range b.Activities.Manufacturing.Products {
			products = append(products, product.TypeID)
		}
	} else if b.Activities.Reaction != nil {
		for _, product := range b.Activities.Reaction.Products {
			products = append(products, product.TypeID)
		}
	} else {
		return nil, nil
	}

	return products, nil
}

func (b *Blueprint) IsReaction() bool {
	return b.Activities.Reaction != nil
}

func (b Blueprint) CanBeBuilt() bool {
	return b.Activities.Manufacturing != nil || b.Activities.Manufacturing != nil
}

func ImportBlueprints(blueprintsFileContents []byte) (ProductBlueprintMap, error) {
	m := make(map[interface{}]interface{})
	yaml.Unmarshal(blueprintsFileContents, &m)

	blueprintMap := make(map[eveonline.TypeID]Blueprint)
	err := yaml.Unmarshal(blueprintsFileContents, &blueprintMap)
	if err != nil {
		return nil, err
	}

	blueprints := make(ProductBlueprintMap)
	for _, blueprint := range blueprintMap {
		products, err := blueprint.CreatesProducts()
		if err != nil {
			return nil, err
		}
		for _, productTypeID := range products {
			blueprintsCreatingProduct, ok := blueprints[productTypeID]
			if !ok {
				blueprintsCreatingProduct = make([]*Blueprint, 0, 1)
			}
			blueprintCopy := new(Blueprint)
			*blueprintCopy = blueprint
			blueprintsCreatingProduct = append(blueprintsCreatingProduct, blueprintCopy)
			blueprints[productTypeID] = blueprintsCreatingProduct
		}
	}

	return blueprints, nil
}
