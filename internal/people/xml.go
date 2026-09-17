package people

import (
	"encoding/xml"
	"fmt"
	"io"
)

const indent = "    "

// WriteXML serialiserar modellen till w som indenterad XML.
func WriteXML(w io.Writer, people *People) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", indent)

	if err := enc.Encode(people); err != nil {
		return fmt.Errorf("kunde inte skriva XML: %w", err)
	}

	if err := enc.Close(); err != nil {
		return err
	}

	_, err := io.WriteString(w, "\n")
	return err
}