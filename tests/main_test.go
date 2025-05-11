package tests

import (
	"testing"

	"go.redsock.ru/evon"
)

func TestMain(m *testing.M) {
	m.Run()
}

type testObject struct {
	RootIntValue  int               `evon:"ROOT-INT-VALUE"`
	ChildObjValue childTestObject   `evon:"CHILD-OBJ-VALUE"`
	Children      []childTestObject `evon:"CHILDREN"`
}

type childTestObject struct {
	StringVal   string `evon:"STRING-VALUE"`
	BoolValFlag bool   `evon:"BOOL-VALUE"`
}

func (c childTestObject) MarshalEnv(prefix string) ([]*evon.Node, error) {
	return []*evon.Node{
		{
			Name:  prefix + "_" + c.StringVal + "-STRING-VALUE",
			Value: c.StringVal,
		},
		{
			Name:  prefix + "_" + c.StringVal + "-BOOL-VALUE",
			Value: c.BoolValFlag,
		},
	}, nil
}

func newTestObject() testObject {
	return testObject{
		RootIntValue: 3,
		ChildObjValue: childTestObject{
			StringVal:   "12",
			BoolValFlag: true,
		},
		Children: []childTestObject{
			{
				StringVal:   "42",
				BoolValFlag: true,
			},
			{
				StringVal:   "52",
				BoolValFlag: false,
			},
		},
	}
}

func expectedObjNodes() *evon.Node {
	to := newTestObject()
	return &evon.Node{
		Name: "",
		InnerNodes: []*evon.Node{
			{
				Name:  "ROOT-INT-VALUE",
				Value: to.RootIntValue,
			},
			{
				Name: "CHILD-OBJ-VALUE",
				InnerNodes: []*evon.Node{
					{
						Name:  "CHILD-OBJ-VALUE_STRING-VALUE",
						Value: to.ChildObjValue.StringVal,
					},
					{
						Name:  "CHILD-OBJ-VALUE_BOOL-VALUE",
						Value: to.ChildObjValue.BoolValFlag,
					},
				},
			},
			{
				Name: "CHILDREN",
				InnerNodes: []*evon.Node{
					{
						Name:  "CHILDREN_42-STRING-VALUE",
						Value: to.Children[0].StringVal,
					},
					{
						Name:  "CHILDREN_42-BOOL-VALUE",
						Value: to.Children[0].BoolValFlag,
					},
					{
						Name:  "CHILDREN_52-STRING-VALUE",
						Value: to.Children[1].StringVal,
					},
					{
						Name:  "CHILDREN_52-BOOL-VALUE",
						Value: to.Children[1].BoolValFlag,
					},
				},
			},
		},
	}
}
