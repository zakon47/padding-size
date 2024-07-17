package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"testing"
)

func TestAnalyzeStruct(t *testing.T) {
	s := &StructInfo{
		Name: "TestStruct",
		Fields: []FieldInfo{
			{Name: "Field1", Type: "int8"},
			{Name: "Field2", Type: "int32"},
			{Name: "Field3", Type: "int16"},
			{Name: "Field4", Type: "int64"},
		},
	}

	analyzeStruct(s)

	expectedSizes := []int64{1, 4, 2, 8}
	expectedAligns := []int64{1, 4, 2, 8}
	expectedOffsets := []int64{0, 4, 8, 16}

	for i, field := range s.Fields {
		if field.Size != expectedSizes[i] {
			t.Errorf("Field %s: expected size %d, got %d", field.Name, expectedSizes[i], field.Size)
		}
		if field.Align != expectedAligns[i] {
			t.Errorf("Field %s: expected align %d, got %d", field.Name, expectedAligns[i], field.Align)
		}
		if field.Offset != expectedOffsets[i] {
			t.Errorf("Field %s: expected offset %d, got %d", field.Name, expectedOffsets[i], field.Offset)
		}
	}

	if s.Size != 24 {
		t.Errorf("Expected struct size 24, got %d", s.Size)
	}
	if s.Align != 8 {
		t.Errorf("Expected struct align 8, got %d", s.Align)
	}
}

func TestOptimizeStruct(t *testing.T) {
	s := &StructInfo{
		Name: "TestStruct",
		Fields: []FieldInfo{
			{Name: "Field1", Type: "int8"},
			{Name: "Field2", Type: "int64"},
			{Name: "Field3", Type: "int32"},
			{Name: "Field4", Type: "int16"},
		},
	}

	fmt.Println("Before optimization:")
	for _, f := range s.Fields {
		fmt.Printf("Field %s: type=%s, size=%d, align=%d, offset=%d\n", f.Name, f.Type, f.Size, f.Align, f.Offset)
	}

	optimizeStruct(s)

	fmt.Println("\nAfter optimization:")
	for _, f := range s.Fields {
		fmt.Printf("Field %s: type=%s, size=%d, align=%d, offset=%d\n", f.Name, f.Type, f.Size, f.Align, f.Offset)
	}

	expectedOrder := []string{"Field2", "Field3", "Field4", "Field1"}
	expectedOffsets := []int64{0, 8, 12, 14}
	expectedSizes := []int64{8, 4, 2, 1}
	expectedAligns := []int64{8, 4, 2, 1}

	for i, field := range s.Fields {
		if field.Name != expectedOrder[i] {
			t.Errorf("Expected field %s at position %d, got %s", expectedOrder[i], i, field.Name)
		}
		if field.Offset != expectedOffsets[i] {
			t.Errorf("Field %s: expected offset %d, got %d", field.Name, expectedOffsets[i], field.Offset)
		}
		if field.Size != expectedSizes[i] {
			t.Errorf("Field %s: expected size %d, got %d", field.Name, expectedSizes[i], field.Size)
		}
		if field.Align != expectedAligns[i] {
			t.Errorf("Field %s: expected align %d, got %d", field.Name, expectedAligns[i], field.Align)
		}
	}

	if s.Size != 16 {
		t.Errorf("Expected optimized struct size 16, got %d", s.Size)
	}

	if s.Align != 8 {
		t.Errorf("Expected struct alignment 8, got %d", s.Align)
	}
}

func TestProcessFile(t *testing.T) {
	src := `
package test

type TestStruct struct {
	Field1 bool ` + "`json:\"field1\"`" + `
	Field2 int32 ` + "`json:\"field2\"`" + `
	Field3 int16 ` + "`json:\"field3\"`" + `
	Field4 int64 ` + "`json:\"field4\"`" + `
}
`

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse test file: %v", err)
	}

	var structs []StructInfo
	ast.Inspect(f, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		structInfo := StructInfo{Name: typeSpec.Name.Name}

		for _, field := range structType.Fields.List {
			fieldType := field.Type.(*ast.Ident).Name
			tag := ""
			if field.Tag != nil {
				tag = field.Tag.Value
			}
			for _, name := range field.Names {
				structInfo.Fields = append(structInfo.Fields, FieldInfo{
					Name: name.Name,
					Type: fieldType,
					Tag:  tag,
				})
			}
		}

		analyzeStruct(&structInfo)
		structs = append(structs, structInfo)
		return true
	})

	if len(structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(structs))
	}

	s := structs[0]
	if s.Name != "TestStruct" {
		t.Errorf("Expected struct name TestStruct, got %s", s.Name)
	}

	expectedFields := []FieldInfo{
		{Name: "Field1", Type: "bool", Tag: "`json:\"field1\"`", Size: 1, Align: 1, Offset: 0},
		{Name: "Field2", Type: "int32", Tag: "`json:\"field2\"`", Size: 4, Align: 4, Offset: 4},
		{Name: "Field3", Type: "int16", Tag: "`json:\"field3\"`", Size: 2, Align: 2, Offset: 8},
		{Name: "Field4", Type: "int64", Tag: "`json:\"field4\"`", Size: 8, Align: 8, Offset: 16},
	}

	if !reflect.DeepEqual(s.Fields, expectedFields) {
		t.Errorf("Fields do not match expected. Got %+v, want %+v", s.Fields, expectedFields)
	}

	if s.Size != 24 {
		t.Errorf("Expected struct size 24, got %d", s.Size)
	}

	if s.Align != 8 {
		t.Errorf("Expected struct align 8, got %d", s.Align)
	}
}
