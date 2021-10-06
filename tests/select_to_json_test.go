package tests

import (
	"cloud.google.com/go/spanner"
	"context"
	"encoding/json"
	"fmt"
	"github.com/10antz-inc/ssorm"
	"google.golang.org/api/iterator"
	"strings"
	"testing"
)

type ColumnTable struct {
	ColumnName  string
	SpannerType string
}

func TestSelectToJson(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"

	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	var columnTable []*ColumnTable

	rtx := client.ReadOnlyTransaction()
	defer rtx.Close()

	err := ssorm.SimpleQueryRead(ctx, rtx, `SELECT t.column_name as ColumnName, t.spanner_type as SpannerType, FROM information_schema.columns AS t WHERE t.table_name = "Singers"`, &columnTable)

	dataTypes := make(map[string]string)
	for i := 0; i < len(columnTable); i++ {
		columnName := columnTable[i].ColumnName
		dataType := columnTable[i].SpannerType
		dataTypes[columnName] = dataType
	}

	stmt := spanner.NewStatement("select * from singers limit 10")
	iter := rtx.Query(ctx, stmt)
	values := readRows(iter)
	result := extractDataByType(dataTypes, values)
	bytes, err := json.Marshal(result)
	if err != nil {
		fmt.Println("JSON marshal error: ", err)
	}

	fmt.Println(string(bytes))

	if err != nil {
		t.Fatalf("Error happened when search singers, got %v", err)
	}
}

func TestAutoGenerateOrm(t *testing.T) {
	url := "projects/spanner-emulator/instances/test/databases/test"

	ctx := context.Background()

	client, _ := spanner.NewClient(ctx, url)
	defer client.Close()

	var columnTable []*ColumnTable

	rtx := client.ReadOnlyTransaction()
	defer rtx.Close()

	tableName := "Singers"
	query := fmt.Sprintf("SELECT t.column_name as ColumnName, t.spanner_type as SpannerType, FROM information_schema.columns AS t WHERE t.table_name = \"%s\"  order by t.ordinal_position", tableName)
	err := ssorm.SimpleQueryRead(ctx, rtx, query, &columnTable)

	dataTypes := make(map[string]string)
	columnNames := []string{}
	for i := 0; i < len(columnTable); i++ {
		columnName := columnTable[i].ColumnName
		dataType := columnTable[i].SpannerType
		dataTypes[columnName] = dataType
		columnNames = append(columnNames, columnName)
	}

	generate := fmt.Sprintf("type %s struct {\n", tableName)
	result := extractColumn(dataTypes)
	for i := 0; i < len(columnNames); i++ {
		k := columnNames[i]
		v := result[k]
		generate += fmt.Sprintf("\t%s %s\n", k, v)
	}
	generate += "}"
	fmt.Println(generate)

	//type Singer struct {
	//	SingerId        int64  `ssorm_key:"primary"`
	//	FirstName       string `spanner:"Name"`
	//	LastName        spanner.NullString
	//	LastName2       spanner.NullString
	//	TestTime        spanner.NullTime     `spanner:"TestTime"` //NULL を許容する場合必ず、spanner.NullTimeを指定すること
	//	TestSpannerTime spanner.NullTime     `spanner:"TestSpannerTime"`
	//	TagIDs          []spanner.NullString `spanner:"TagIds"`
	//	Numbers         []int64              `spanner:"Numbers"`
	//	Albums          []*Albums            `spanner:"Albums"`
	//	Concerts        []*Concerts          `spanner:"Concerts"`
	//	DeleteTime      spanner.NullTime     `spanner:"DeleteTime" ssorm_key:"delete_time"`
	//	CreateTime      time.Time            `spanner:"CreateTime" ssorm_key:"create_time"`
	//	UpdateTime      time.Time            `spanner:"UpdateTime" ssorm_key:"update_time"`
	//}

	if err != nil {
		t.Fatalf("Error happened when search singers, got %v", err)
	}
}

func readRows(iter *spanner.RowIterator) []spanner.Row {
	var rows []spanner.Row
	defer iter.Stop()
	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			//log.Println("Failed to read data, err = %s", err)
		}
		rows = append(rows, *row)
	}
	return rows
}

func decodeValueByType(index int, row spanner.Row, value interface{}) {
	err := row.Column(index, value)
	if err != nil {
		//log.Println("Failed to extract value, err = %s", err)
	}
}

func extractDataByType(types map[string]string, rows []spanner.Row) []*map[string]interface{} {
	result := make([]*map[string]interface{}, len(rows))
	index := 0
	for _, row := range rows {
		valueMap := make(map[string]interface{})
		columnNames := row.ColumnNames()
		for i := 0; i < row.Size(); i++ {
			if strings.Index(types[columnNames[i]], "STRING") == 0 {
				var value spanner.NullString
				decodeValueByType(i, row, &value)
				valueMap[columnNames[i]] = value
				continue
			}
			if strings.Index(types[columnNames[i]], "ARRAY") == 0 {
				if strings.Index(types[columnNames[i]], "STRING") > 0 {
					var value []spanner.NullString
					decodeValueByType(i, row, &value)
					if value != nil {
						var strValue []string
						for _, v := range value {
							strValue = append(strValue, fmt.Sprintf(`"%v"`, v.StringVal))
						}
						valueMap[columnNames[i]] = "[" + strings.Join(strValue, ",") + "]"
					} else {
						valueMap[columnNames[i]] = fmt.Sprintf("%v", value)
					}
					continue

				}
				if strings.Index(types[columnNames[i]], "INT") > 0 {
					var value []spanner.NullInt64
					decodeValueByType(i, row, &value)
					if value != nil {
						var strValue []string
						for _, v := range value {
							strValue = append(strValue, fmt.Sprintf("%v", v.Int64))
						}

						valueMap[columnNames[i]] = "[" + strings.Join(strValue, ",") + "]"
					} else {
						valueMap[columnNames[i]] = fmt.Sprintf("%v", value)
					}
				}
				if strings.Index(types[columnNames[i]], "FLOAT") > 0 {
					var value []spanner.NullFloat64
					decodeValueByType(i, row, &value)
					if value != nil {
						var strValue []string
						for _, v := range value {
							strValue = append(strValue, fmt.Sprintf("%v", v.Float64))
						}
						valueMap[columnNames[i]] = "[" + strings.Join(strValue, ",") + "]"
					} else {
						valueMap[columnNames[i]] = fmt.Sprintf("%v", value)
					}
				}

				continue
			}
			switch types[columnNames[i]] {
			case "TIMESTAMP":
				var value spanner.NullTime
				decodeValueByType(i, row, &value)
				valueMap[columnNames[i]] = value
			case "INT64":
				var value spanner.NullInt64
				decodeValueByType(i, row, &value)
				valueMap[columnNames[i]] = value
			case "FLOAT64":
				var value spanner.NullFloat64
				decodeValueByType(i, row, &value)
				valueMap[columnNames[i]] = value
			case "BOOL":
				var value spanner.NullBool
				decodeValueByType(i, row, &value)
				valueMap[columnNames[i]] = value
			}
		}
		result[index] = &valueMap
		index++
	}
	//log.Println("parquet format: %s", md)
	return result
}

func extractColumn(types map[string]string) map[string]string {

	valueMap := make(map[string]string)
	var (
		dataType   string
		spannerTag string
	)
	for k, v := range types {
		ormTag := ""
		spannerTag = fmt.Sprintf("%s:\"%s\"", "spanner", k)
		if strings.Index(v, "STRING") == 0 {
			dataType = "spanner.NullString"
		}
		if strings.Index(v, "ARRAY") == 0 {
			if strings.Index(v, "STRING") > 0 {
				dataType = "[]spanner.NullString"
			}
			if strings.Index(v, "INT") > 0 {
				dataType = "[]spanner.NullInt64"
			}
			if strings.Index(v, "FLOAT") > 0 {
				dataType = "[]spanner.NullFloat64"
			}

		}

		switch v {
		case "TIMESTAMP":
			dataType = "spanner.NullTime"
		case "INT64":
			dataType = "spanner.NullInt64"
		case "FLOAT64":
			dataType = "spanner.NullFloat64"
		case "BOOL":
			dataType = "bool"
		}

		switch k {
		case "CreateTime":
			ormTag = "ssorm_key:\"create_time\""
		case "UpdateTime":
			ormTag = "ssorm_key:\"update_time\""
		case "DeleteTime":
			ormTag = "ssorm_key:\"delete_time\""

		}
		if ormTag == "" {
			valueMap[k] = fmt.Sprintf("%s `%s`", dataType, spannerTag)
		} else {
			valueMap[k] = fmt.Sprintf("%s `%s %s`", dataType, spannerTag, ormTag)
		}
	}

	//log.Println("parquet format: %s", md)
	return valueMap
}