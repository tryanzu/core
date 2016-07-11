package components

import (
	"github.com/fernandez14/spartangeek-blacker/modules/exceptions"
	"github.com/fernandez14/spartangeek-blacker/modules/search"
	"github.com/fernandez14/spartangeek-blacker/mongo"
	"gopkg.in/mgo.v2/bson"
)

func Boot() *Module {

	module := &Module{}

	return module
}

type Module struct {
	Mongo  *mongo.Service `inject:""`
	Search *search.Module `inject:""`
}

func (module *Module) Get(find interface{}) (*ComponentModel, error) {

	var model interface{}

	context := module
	database := module.Mongo.Database

	switch find.(type) {
	case bson.ObjectId:

		// Get the user using it's id
		err := database.C("components").FindId(find.(bson.ObjectId)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid component id. Not found."}
		}

	case bson.M:

		// Get the user using it's id
		err := database.C("components").Find(find.(bson.M)).One(&model)

		if err != nil {

			return nil, exceptions.NotFound{"Invalid component finder. Not found."}
		}

	case *ComponentModel:

		component := find.(*ComponentModel)
		component.SetDI(context)

		return component, nil

	default:
		panic("Unkown argument")
	}

	// Marshal the data inside the generic model
	encoded, err := bson.Marshal(model)

	if err != nil {
		panic(err)
	}

	var component *ComponentModel

	err = bson.Unmarshal(encoded, &component)

	if err != nil {
		panic(err)
	}

	component.SetDI(context)
	component.SetGeneric(encoded)

	return component, nil
}

func (module *Module) List(limit, offset int, search, kind string, activated bool) ([]Component, []ComponentTypeCountModel, int) {

	components := make([]Component, 0)
	facets := make([]ComponentTypeCountModel, 0)
	database := module.Mongo.Database

	// Fields to retrieve
	fields := ComponentFields
	query := bson.M{}

	if kind != "" {
		query["type"] = kind
	}

	if activated == true {
		query["activated"] = true
		query["store.vendors.spartangeek.stock"] = bson.M{"$ne": 0}
	}

	if search != "" {
		fields["score"] = bson.M{"$meta": "textScore"}
		query["$text"] = bson.M{"$search": search}
	}

	if _, exists := query["$text"]; exists {

		err := database.C("components").Find(query).Select(fields).Sort("$textScore:score").Limit(limit).Skip(offset).All(&components)

		if err != nil {
			panic(err)
		}
	} else {

		err := database.C("components").Find(query).Select(fields).Limit(limit).Sort("-views", "-activated").Skip(offset).All(&components)

		if err != nil {
			panic(err)
		}
	}

	rows, err := database.C("components").Find(query).Count()

	if err != nil {
		panic(err)
	}

	// Remove type from aggregation since its not needed
	delete(query, "type")

	err = database.C("components").Pipe([]bson.M{
		{"$match": query},
		{"$group": bson.M{"_id": "$type", "count": bson.M{"$sum": 1}}},
	}).All(&facets)

	if err != nil {
		panic(err)
	}

	return components, facets, rows
}
