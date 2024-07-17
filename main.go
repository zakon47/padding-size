package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FieldInfo represents information about a struct field
type FieldInfo struct {
	Name    string
	Type    string
	Tag     string
	Size    int64
	Align   int64
	Offset  int64
	Comment *ast.CommentGroup
}

// StructInfo represents information about a struct
type StructInfo struct {
	Name   string
	Fields []FieldInfo
	Size   int64
	Align  int64
}

func main() {
	fix := flag.Bool("fix", false, "Apply fixes to optimize struct layout")
	help := flag.Bool("help", false, "Display help information")
	flag.Parse()

	if *help || len(os.Args) == 1 {
		printHelp()
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Error: No input files or directories specified.")
		fmt.Println("Run 'padding-size -help' for usage information.")
		os.Exit(1)
	}

	for _, path := range args {
		err := processPath(path, *fix)
		if err != nil {
			fmt.Printf("Error processing %s: %v\n", path, err)
		}
	}
}

func printHelp() {
	fmt.Println("padding-size - Analyze and optimize struct field alignment in Go")
	fmt.Println("\nUsage:")
	fmt.Println("  padding-size [options] <file or directory paths>")
	fmt.Println("\nOptions:")
	fmt.Println("  -fix        Apply fixes to optimize struct layout")
	fmt.Println("  -help       Display this help information")
	fmt.Println("\nExamples:")
	fmt.Println("  padding-size main.go")
	fmt.Println("  padding-size -fix .")
	fmt.Println("  padding-size -fix /path/to/project")
}

func processPath(path string, fix bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return filepath.Walk(path, func(filePath string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !fileInfo.IsDir() && strings.HasSuffix(filePath, ".go") {
				return processFile(filePath, fix)
			}
			return nil
		})
	}

	return processFile(path, fix)
}

func processFile(filePath string, fix bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	var structs []StructInfo

	ast.Inspect(node, func(n ast.Node) bool {
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
			fieldType := types.ExprString(field.Type)
			tag := ""
			if field.Tag != nil {
				tag = field.Tag.Value
			}
			for _, name := range field.Names {
				structInfo.Fields = append(structInfo.Fields, FieldInfo{
					Name:    name.Name,
					Type:    fieldType,
					Tag:     tag,
					Comment: field.Comment,
				})
			}
		}

		analyzeStruct(&structInfo)
		structs = append(structs, structInfo)
		return true
	})

	if len(structs) > 0 {
		fmt.Printf("File: %s\n", filePath)
		for _, s := range structs {
			printStructInfo(s)
			if fix {
				optimizeStruct(&s)
				printStructInfo(s)
			}
		}

		if fix {
			return applyFixes(filePath, structs, fset, node)
		}
	}

	return nil
}

func analyzeStruct(s *StructInfo) {
	var offset int64
	var maxAlign int64 = 1
	for i := range s.Fields {
		s.Fields[i].Size = getFieldSize(s.Fields[i].Type)
		s.Fields[i].Align = getFieldAlign(s.Fields[i].Type)
		if s.Fields[i].Align > maxAlign {
			maxAlign = s.Fields[i].Align
		}
		offset = align(offset, s.Fields[i].Align)
		s.Fields[i].Offset = offset
		offset += s.Fields[i].Size
	}
	s.Size = align(offset, maxAlign)
	s.Align = maxAlign
}

func getFieldSize(fieldType string) int64 {
	switch fieldType {
	case "bool", "int8", "uint8", "byte":
		return 1
	case "int16", "uint16":
		return 2
	case "int32", "uint32", "float32":
		return 4
	case "int64", "uint64", "float64", "complex64":
		return 8
	case "string", "[]byte", "[]rune", "error", "complex128":
		return 16 // Assuming 64-bit architecture (8 bytes for pointer, 8 for length)
	default:
		if strings.HasPrefix(fieldType, "*") {
			return 8 // Assuming 64-bit architecture
		}
		// For other types (structs, arrays, etc.), we need more sophisticated analysis
		// For simplicity, we'll assume 8 bytes, but this should be improved
		return 8
	}
}

func getFieldAlign(fieldType string) int64 {
	switch fieldType {
	case "bool", "int8", "uint8", "byte":
		return 1
	case "int16", "uint16":
		return 2
	case "int32", "uint32", "float32":
		return 4
	default:
		// For most types on 64-bit systems, alignment is 8
		return 8
	}
}

func align(offset, align int64) int64 {
	return (offset + align - 1) &^ (align - 1)
}

func printStructInfo(s StructInfo) {
	fmt.Printf("Struct: %s (size: %d bytes, align: %d)\n", s.Name, s.Size, s.Align)
	for _, field := range s.Fields {
		fmt.Printf("  %s %s (offset: %d, size: %d, align: %d)\n",
			field.Name, field.Type, field.Offset, field.Size, field.Align)
	}
	fmt.Println()
}

func optimizeStruct(s *StructInfo) {
	// First, analyze the struct to set correct sizes and alignments
	analyzeStruct(s)

	// Now sort the fields
	sort.Slice(s.Fields, func(i, j int) bool {
		if s.Fields[i].Align != s.Fields[j].Align {
			return s.Fields[i].Align > s.Fields[j].Align
		}
		return s.Fields[i].Size > s.Fields[j].Size
	})

	// Recalculate offsets after sorting
	var offset int64
	var maxAlign int64 = 1
	for i := range s.Fields {
		if s.Fields[i].Align > maxAlign {
			maxAlign = s.Fields[i].Align
		}
		offset = align(offset, s.Fields[i].Align)
		s.Fields[i].Offset = offset
		offset += s.Fields[i].Size
	}
	s.Size = align(offset, maxAlign)
	s.Align = maxAlign
}

func applyFixes(filePath string, structs []StructInfo, fset *token.FileSet, node *ast.File) error {
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		for _, s := range structs {
			if typeSpec.Name.Name == s.Name {
				newFields := make([]*ast.Field, len(s.Fields))
				for i, field := range s.Fields {
					newFields[i] = &ast.Field{
						Names: []*ast.Ident{ast.NewIdent(field.Name)},
						Type:  ast.NewIdent(field.Type),
					}
					if field.Tag != "" {
						newFields[i].Tag = &ast.BasicLit{
							Kind:  token.STRING,
							Value: field.Tag,
						}
					}
					if field.Comment != nil {
						newFields[i].Comment = field.Comment
					}
				}
				structType.Fields.List = newFields
				break
			}
		}
		return true
	})

	var buf strings.Builder
	err := format.Node(&buf, fset, node)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(buf.String()), 0644)
}
