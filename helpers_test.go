package evon

import (
	_ "embed"

	"go.redsock.ru/toolbox"
)

var (
	//go:embed tests/object.env
	fullObjectDotEnv []byte
	//go:embed tests/prefixed-object.env
	prefixedExpectedDotEnv []byte
	//go:embed tests/simple-object.env
	simpleObjectDotEnv []byte
	//go:embed tests/matreshka.env
	matreshkaDotEnv []byte

	//go:embed tests/complex-config.yaml
	complexYamlConfig []byte
)

type TestObject struct {
	RootIntValue  int               `evon:"ROOT-INT-VALUE"`
	ChildObjValue ChildTestObject   `evon:"CHILD-OBJ-VALUE"`
	PointerValue  *int              `evon:"POINTER-VALUE"`
	Children      []ChildTestObject `evon:"CHILDREN"`
	UnknownArray  []UntaggedStruct  `evon:"UNKNOWN-ARRAY"`
}

type UntaggedStruct struct {
	A int
}

type ChildTestObject struct {
	StringVal   string `evon:"STRING-VALUE"`
	BoolValFlag bool   `evon:"BOOL-VALUE"`
}

func (c ChildTestObject) MarshalEnv(prefix string) ([]*Node, error) {
	return []*Node{
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

func NewTestObject() TestObject {
	return TestObject{
		RootIntValue: 3,
		PointerValue: toolbox.ToPtr(3),
		ChildObjValue: ChildTestObject{
			StringVal:   "12",
			BoolValFlag: true,
		},
		Children: []ChildTestObject{
			{
				StringVal:   "42",
				BoolValFlag: true,
			},
			{
				StringVal:   "52",
				BoolValFlag: false,
			},
		},
		UnknownArray: []UntaggedStruct{{A: 1}},
	}
}

func (t *TestObject) ExpectedObjNodes(prefix string) *Node {
	to := NewTestObject()

	var rootPref string
	if prefix != "" {
		rootPref = prefix
		prefix += "_"
	}

	out := &Node{
		Name: rootPref,
		InnerNodes: []*Node{
			{
				Name:  prefix + "ROOT-INT-VALUE",
				Value: to.RootIntValue,
			},
			{
				Name: prefix + "CHILD-OBJ-VALUE",
				InnerNodes: []*Node{
					{
						Name:  prefix + "CHILD-OBJ-VALUE_STRING-VALUE",
						Value: to.ChildObjValue.StringVal,
					},
					{
						Name:  prefix + "CHILD-OBJ-VALUE_BOOL-VALUE",
						Value: to.ChildObjValue.BoolValFlag,
					},
				},
			},
		},
	}

	if t.PointerValue != nil {
		out.InnerNodes = append(out.InnerNodes, &Node{
			Name:  prefix + "POINTER-VALUE",
			Value: *to.PointerValue,
		})
	}

	out.InnerNodes = append(out.InnerNodes,
		&Node{
			Name: prefix + "CHILDREN",
			InnerNodes: []*Node{
				{
					Name:  prefix + "CHILDREN_42-STRING-VALUE",
					Value: to.Children[0].StringVal,
				},
				{
					Name:  prefix + "CHILDREN_42-BOOL-VALUE",
					Value: to.Children[0].BoolValFlag,
				},
				{
					Name:  prefix + "CHILDREN_52-STRING-VALUE",
					Value: to.Children[1].StringVal,
				},
				{
					Name:  prefix + "CHILDREN_52-BOOL-VALUE",
					Value: to.Children[1].BoolValFlag,
				},
			},
		})

	out.InnerNodes = append(out.InnerNodes,
		&Node{
			Name: prefix + "UNKNOWN-ARRAY",
			InnerNodes: []*Node{
				{
					Name: prefix + "UNKNOWN-ARRAY_[0]",
					InnerNodes: []*Node{
						{
							Name:  prefix + "UNKNOWN-ARRAY_[0]_A",
							Value: to.UnknownArray[0].A,
						},
					},
				},
			},
		})

	return out
}
