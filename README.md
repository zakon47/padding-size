# padding-size

`padding-size` is a Go tool that analyzes and optimizes struct field alignment in Go source files. It helps developers identify and fix inefficient memory layouts in structs, potentially reducing memory usage and improving performance.

## Features

- Analyzes struct field alignment and padding in Go source files
- Calculates the size and alignment of each struct and its fields
- Optimizes struct layout by reordering fields for better memory efficiency
- Supports processing single files or entire directories
- Provides an option to automatically apply optimizations to source files

## Installation

To install `padding-size`, make sure you have Go installed on your system, then run:

```
go get -u github.com/yourusername/padding-size
```

Replace `yourusername` with your actual GitHub username.

## Usage

```
padding-size [options] <file or directory paths>
```

### Options

- `-fix`: Apply fixes to optimize struct layout
- `-help`: Display help information

### Examples

Analyze a single file:
```
padding-size main.go
```

Optimize structs in all Go files in the current directory:
```
padding-size -fix .
```

Analyze all Go files in a specific directory:
```
padding-size /path/to/project
```

## Output

For each struct found in the processed files, `padding-size` will output:

- Struct name
- Total size of the struct
- Alignment of the struct
- For each field:
    - Field name
    - Field type
    - Offset within the struct
    - Size of the field
    - Alignment of the field

If the `-fix` option is used, it will also show the optimized layout of the struct.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
