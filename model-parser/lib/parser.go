package lib

import (
	"backend/model-parser/model"
	"github.com/google/uuid"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	data     model.ProcessData
	document model.DocNode
}

func (parser *Parser) Parse(data model.ProcessData) *model.Record {
	parser.data = data

	processor := new(Processor)

	// parse html
	doc, err := ParseHtml(*data.Html)

	if err != nil {
		log.Panic(err)
	}

	pageData := model.PageData{
		Document: doc,
		Url:      *data.Url,
	}

	docNode := processor.ProcessWithModel(data.Model, &pageData)
	parser.document = docNode

	return parser.ParseWithDocNode(docNode)
}

func (parser Parser) ParseWithDocNode(processedDocument model.DocNode) *model.Record {
	recordData := parser.Extract(parser.data.Schema, processedDocument)
	recordMeta := parser.ExtractMeta(true)

	record := new(model.Record)
	record.Data = recordData
	record.Meta = recordMeta
	record.Tags = append([]string{})

	record.Schema = parser.data.Model.Schema
	record.Source = parser.data.Model.Source
	record.SourceUrl = *parser.data.Url
	record.Ref = parser.extractRef(parser.data.Model, *parser.data.Url)
	record.ObjectType = parser.data.Model.ObjectType

	id, err := uuid.Parse(NamedUUID([]byte(*parser.data.Url)))

	check(err)

	record.Id = id

	record.Name = Coalesce(recordData["name"], recordMeta["title"], recordMeta["og:title"])
	record.Description = Coalesce(recordData["description"], recordMeta["description"], recordMeta["og:description"])

	return record
}

func (parser Parser) Extract(objectProperties model.ObjectProperty, parent model.DocNode) model.RecordData {
	var data = make(model.RecordData)

	for key, property := range objectProperties.GetProperties() {
		value := parser.locatePropertyValue(parent, key, &property)

		if value != nil {
			data[key] = value
		}
	}

	return data
}

func (parser Parser) locatePropertyValue(parent model.DocNode, key string, property *model.SchemaProperty) model.Value {
	fields := parent.Find("[ug-field=\"" + key + "\"]")

	return parser.locatePropertyValueForField(property, fields)
}

func (parser Parser) ExtractMeta(onlyUgField bool) model.RecordMeta {
	metaData := make(model.RecordMeta)

	selector := "meta[ug-field]"
	if !onlyUgField {
		selector = "meta"
	}

	metaFields := parser.document.Find(selector)

	for _, metaField := range metaFields {
		key := metaField.Attr("ug-field")
		value := metaField.Attr("ug-value")

		if strings.HasPrefix(key, "meta.") {
			key = key[len("meta."):]
		}

		metaData[key] = value
	}

	return metaData
}

func (parser Parser) locatePropertyValueForField(property *model.SchemaProperty, fields []model.DocNode) model.Value {
	if property.IsArrayProperty() {
		itemsProperty := property.Items

		var result []model.Value

		for _, field := range fields {
			val := parser.locatePropertyValueField(itemsProperty, field)
			result = append(result, val)
		}

		return result
	} else {
		if len(fields) > 1 && property.IsStringProperty() {
			result := new(strings.Builder)

			isFirst := true
			for _, field := range fields {
				val := strings.TrimSpace(parser.locatePropertyValueField(property, field).(string))

				if val == "" {
					continue
				}
				if !isFirst {
					result.WriteString(" ")
				}

				isFirst = false
				result.WriteString(val)
			}

			return result.String()
		} else if len(fields) > 0 {
			return parser.locatePropertyValueField(property, fields[0])
		} else {
			return nil
		}
	}
}

func (parser Parser) locatePropertyValueField(property *model.SchemaProperty, field model.DocNode) model.Value {
	if property.IsStringProperty() {
		return parser.getValue(field)
	} else if property.IsNumberProperty() {
		val := parser.getValue(field)

		valStrNum := string(regexp.MustCompile("[^\\d.]+").ReplaceAll([]byte(val), []byte("")))

		if len(valStrNum) == 0 {
			return nil
		} else {
			number, err := strconv.ParseFloat(valStrNum, 64)

			check(err)

			return number
		}
	} else if property.IsArrayProperty() {
		return parser.locatePropertyValueForField(property, append([]model.DocNode{}, field))
	} else if property.IsObjectProperty() {
		return parser.Extract(property, field)
	} else if property.IsReferenceProperty() {
		return parser.extractReference(property, field)
	} else {
		log.Print("invalid property type: ", property.Type)
		return nil
	}

}

func (parser Parser) extractReference(referenceProperty *model.SchemaProperty, field model.DocNode) *model.Reference {
	reference := new(model.Reference)

	reference.Name = parser.getValue(field)
	if !field.HasAttr("href") {
		return reference
	}

	href := field.Attr("href")

	if !strings.HasPrefix(href, "http") {
		if strings.HasPrefix(href, "/") {
			pageUrlObj, err := url.Parse(*parser.data.Url)

			check(err)

			href = pageUrlObj.Scheme + "://" + pageUrlObj.Host + href
		}
	}

	schemaName := referenceProperty.Schema

	m := parser.locateModel(schemaName)

	if m == nil {
		return reference
	}

	reference.Source = m.Source
	reference.SourceUrl = href
	reference.Ref = parser.extractRef(m, href)
	reference.ObjectType = m.ObjectType

	return reference
}

func (parser Parser) getValue(field model.DocNode) string {
	if field.HasAttr("ug-value") {
		return strings.TrimSpace(field.Attr("ug-value"))
	}

	return strings.TrimSpace(field.Text())
}

func (parser Parser) getText(selector string) string {
	if parser.document.FindSingle(selector) == nil {
		return ""
	}
	return parser.document.FindSingle(selector).Text()
}

func (parser *Parser) locateModel(schemaName string) *model.Model {
	for _, item := range parser.data.AdditionalModels {
		if item.Schema == schemaName {
			return &item
		}
	}

	return nil
}

func (parser *Parser) extractRef(m *model.Model, href string) string {
	ref := m.Ref

	if ref != "" {
		refExp, err := regexp.Compile(ref)

		check(err)

		h := refExp.FindAllSubmatchIndex([]byte(href), -1)

		if len(h) > 0 {
			return href[h[0][2]:h[0][3]]
		}
	}

	return ""
}

func (parser *Parser) ParseStaticData(p model.ProcessDataLight) (*model.Record, error) {
	record := new(model.Record)

	record.Schema = "dynamic"
	record.Source = p.PageRef.WebsiteName
	record.SourceUrl = p.PageRef.Url
	record.Ref = p.PageRef.Url
	record.ObjectType = "dynamic"
	record.Tags = p.PageRef.Tags

	// parse html
	doc, err := ParseHtml(*p.Html)

	if err != nil {
		log.Panic(err)
	}

	parser.document = doc

	record.Meta = parser.ExtractMeta(false)

	record.Data = parser.ExtractLinks()
	record.Data["title"] = parser.getText("title")

	return record, nil
}

func (parser *Parser) ExtractLinks() model.RecordData {
	linkFields := parser.document.Find("a[href]")

	data := model.RecordData{}

	type LinkData struct {
		Href string `json:"href"`
		Text string `json:"text"`
	}

	var links []LinkData

	for _, linkField := range linkFields {
		href := linkField.Attr("href")
		text := linkField.Text()

		links = append(links, LinkData{
			Href: href,
			Text: text,
		})
	}

	data["links"] = links

	return data
}
