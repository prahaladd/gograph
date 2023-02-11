package omg

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/prahaladd/gograph/core"
	"github.com/stretchr/testify/suite"
)

type MapperTestSuite struct {
	suite.Suite
	mapper Mapper
}

func (suite *MapperTestSuite) TestMapVertexForStructWithReflectionNoLabels() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	suite.Equal("person", v.Labels[0])
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
}

func (suite *MapperTestSuite) TestMapVertexForStructWithReflectionLabels() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(ts, []string{"TestLabel1", "TestLabel2"})
	suite.NoError(err)
	suite.Equal([]string{"TestLabel1", "TestLabel2"}, v.Labels)
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
}

func (suite *MapperTestSuite) TestMapVertexForStructPtrWithReflectionNoLabels() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(&ts, []string{})
	suite.NoError(err)
	suite.Equal("person", v.Labels[0])
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
}

func (suite *MapperTestSuite) TestMapEdgeForStructWithReflectionNoLabels() {
	suite.mapper = NewReflectionMapper()
	ts := livesin{Since: 1990}
	emptyLabel := ""
	e, err := suite.mapper.ToEdge(ts, &emptyLabel)
	suite.NoError(err)
	suite.Equal("livesin", e.Type)
	props := core.KVMap{"since": int32(1990)}
	suite.Equal(props, e.Properties)
}

func (suite *MapperTestSuite) TestMapEdgeForStructWithReflectionLabelIsNil() {
	suite.mapper = NewReflectionMapper()
	ts := livesin{Since: 1990}

	e, err := suite.mapper.ToEdge(ts, nil)
	suite.NoError(err)
	suite.Equal("livesin", e.Type)
	props := core.KVMap{"since": int32(1990)}
	suite.Equal(props, e.Properties)
}

func (suite *MapperTestSuite) TestMapEdgeForStructWithReflectionWithLabels() {
	suite.mapper = NewReflectionMapper()
	ts := livesin{Since: 1990}
	label := "resides_in"
	e, err := suite.mapper.ToEdge(ts, &label)
	suite.NoError(err)
	suite.Equal("resides_in", e.Type)
	props := core.KVMap{"since": int32(1990)}
	suite.Equal(props, e.Properties)
}

func (suite *MapperTestSuite) TestMapVertexToStructInitialized() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	suite.Equal("person", v.Labels[0])
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
	ts2 := person{}
	suite.mapper.FromVertex(v, &ts2)
	suite.Equal(ts2, ts)
}

func (suite *MapperTestSuite) TestMapVertexToStructUnInitialized() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	suite.Equal("person", v.Labels[0])
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
	var ts2 person
	suite.mapper.FromVertex(v, &ts2)
	suite.Equal(ts2, ts)
}

func (suite *MapperTestSuite) TestMapVertexToStructIncorrectPointerType() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	suite.Equal("person", v.Labels[0])
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
	var ts2 int32
	err = suite.mapper.FromVertex(v, &ts2)
	suite.Error(err)
}

func (suite *MapperTestSuite) TestMapVertexToStructNoFieldTags() {
	suite.mapper = NewReflectionMapper()
	ts := testVertex{Field1: "Tom", Field2: "Jerry"}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	suite.Equal("testVertex", v.Labels[0])
	props := core.KVMap{"Field1": "Tom", "Field2": "Jerry"}
	suite.Equal(props, v.Properties)
	var ts2 testVertex
	suite.mapper.FromVertex(v, &ts2)
	suite.Equal(ts2, ts)
}

func (suite *MapperTestSuite) TestMapVertexToStructNonPointerToStruct() {
	suite.mapper = NewReflectionMapper()
	ts := person{Name: "Tom", Age: 12, Department: "Dev"}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	suite.Equal("person", v.Labels[0])
	props := core.KVMap{"name": "Tom", "age": int32(12), "dept": "Dev"}
	suite.Equal(props, v.Properties)
	var ts2 person
	err = suite.mapper.FromVertex(v, ts2)
	suite.Error(err)
}

func (suite *MapperTestSuite) TestMapEdgeToStructInitialized() {
	suite.mapper = NewReflectionMapper()
	ts := livesin{Since: 1990}
	emptyLabel := ""
	e, err := suite.mapper.ToEdge(ts, &emptyLabel)
	suite.NoError(err)
	suite.Equal("livesin", e.Type)
	props := core.KVMap{"since": int32(1990)}
	suite.Equal(props, e.Properties)
	ts2 := livesin{}
	err = suite.mapper.FromEdge(e, &ts2)
	suite.NoError(err)
	suite.Equal(ts, ts2)
}

func (suite *MapperTestSuite) TestMapEdgeToStructNotInitialized() {
	suite.mapper = NewReflectionMapper()
	ts := livesin{Since: 1990}
	emptyLabel := ""
	e, err := suite.mapper.ToEdge(ts, &emptyLabel)
	suite.NoError(err)
	suite.Equal("livesin", e.Type)
	props := core.KVMap{"since": int32(1990)}
	suite.Equal(props, e.Properties)
	var ts2 livesin
	err = suite.mapper.FromEdge(e, &ts2)
	suite.NoError(err)
	suite.Equal(ts, ts2)
}

func (suite *MapperTestSuite) TestMapEdgeToStructIncorrectPointerType() {
	suite.mapper = NewReflectionMapper()
	ts := livesin{Since: 1990}
	emptyLabel := ""
	e, err := suite.mapper.ToEdge(ts, &emptyLabel)
	suite.NoError(err)
	suite.Equal("livesin", e.Type)
	props := core.KVMap{"since": int32(1990)}
	suite.Equal(props, e.Properties)
	var ts2 int32
	err = suite.mapper.FromEdge(e, &ts2)
	suite.Error(err)
}

func (suite *MapperTestSuite) TestMapEdgeToStructNoFieldTags() {
	suite.mapper = NewReflectionMapper()
	ts := testEdge{Field1: "testproperty"}
	e, err := suite.mapper.ToEdge(ts, nil)
	suite.NoError(err)
	suite.Equal("testEdge", e.Type)
	props := core.KVMap{"Field1": "testproperty"}
	suite.Equal(props, e.Properties)
	var ts2 testEdge
	suite.mapper.FromEdge(e, &ts2)
	suite.Equal(ts2, ts)
}

func (suite *MapperTestSuite) TestMapEdgeToStructNonPointerToStruct() {
	suite.mapper = NewReflectionMapper()
	ts := testEdge{Field1: "testproperty"}
	e, err := suite.mapper.ToEdge(ts, nil)
	suite.NoError(err)
	suite.Equal("testEdge", e.Type)
	props := core.KVMap{"Field1": "testproperty"}
	suite.Equal(props, e.Properties)
	var ts2 testEdge
	err = suite.mapper.FromEdge(e, ts2)
	suite.Error(err)
}

func (suite *MapperTestSuite) TestPlayReflection() {
	ts := person{Name: "Tom"}
	tsType := reflect.TypeOf(&ts)
	fmt.Println("type: ", tsType.Name(), " kind: ", tsType.Kind())
	if tsType.Kind() == reflect.Ptr {
		fmt.Println("elem type: ", tsType.Elem().Name(), " elem kind: ", tsType.Elem().Kind())
		val := reflect.ValueOf(ts)
		for i := 0; i < val.NumField(); i++ {
			fmt.Println("Field : ", tsType.Elem().Field(i).Name, " type: ", tsType.Elem().Field(i).Type, " type through val: ", val.Field(i).Type(), " Field value:", val.Field(i).Interface(), " Tag:", tsType.Elem().Field(i).Tag)
		}
	}
}

func (suite *MapperTestSuite) TestMapComplexStructToVertex() {
	suite.mapper = NewReflectionMapper()
	ts := nested{Field1: "Hello World", Field2: int32(8), ComplexField: tuple{Field1: "Tom and Jerry", Field2: 15}}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	t, ok := v.Properties["ComplexField"]
	suite.True(ok)
	_, ok = t.(tuple)
	suite.True(ok)

}

func (suite *MapperTestSuite) TestMapVertexToComplexStruct() {
	suite.mapper = NewReflectionMapper()
	ts := nested{Field1: "Hello World", Field2: int32(8), ComplexField: tuple{Field1: "Tom and Jerry", Field2: 15}}
	v, err := suite.mapper.ToVertex(ts, []string{})
	suite.NoError(err)
	t, ok := v.Properties["ComplexField"]
	suite.True(ok)
	_, ok = t.(tuple)
	suite.True(ok)

	// perform a reverse mapper
	var n nested
	err = suite.mapper.FromVertex(v, &n)
	suite.NoError(err)

	suite.Equal(ts, n)

}

func TestMapperTestSuite(t *testing.T) {
	suite.Run(t, new(MapperTestSuite))
}

type person struct {
	Name       string `ogm:"name"`
	Age        int32  `ogm:"age"`
	Department string `ogm:"dept"`
}

type livesin struct {
	Since int32 `ogm:"since"`
}

type testVertex struct {
	Field1 string
	Field2 string
}

type testEdge struct {
	Field1 string
}

type tuple struct {
	Field1 string
	Field2 int32
}

type nested struct {
	Field1       string
	Field2       int32
	ComplexField tuple
}
