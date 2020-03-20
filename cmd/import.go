package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Imports CSV files into InfluxDB",
	Long:  `Imports CSV (query result) file into InfluxDB.`,
	Run: func(cmd *cobra.Command, args []string) {
		fileName, _ := cmd.Flags().GetString("file")
		var reader *csv.Reader
		if len(fileName) > 0 {
			file, err := os.Open(fileName)
			if err != nil {
				log.Fatal(err)
			}
			reader = csv.NewReader(file)
			defer file.Close()
		} else {
			reader = csv.NewReader(os.Stdin)
		}
		processLines(reader)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP("file", "f", "", "The path to the file to import")
}

func processLines(reader *csv.Reader) {
	var table = CsvTable{}
	lineNumber := 0
	for {
		// Read each record from csv
		row, err := reader.Read()
		lineNumber++
		reader.FieldsPerRecord = 0 // since every row can have different items
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if table.AddRow(row) {
			line, err := table.CreateLine(row)
			if err != nil {
				log.Printf("ERROR line #%d: %v\n", lineNumber, err)
			} else {
				fmt.Println(line)
			}
		}
	}
}
