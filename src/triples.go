package main

import (
	"github.com/knakk/rdf"
	"os"
)

func getMatchingTriples(triples []rdf.Triple, subject, predicate, object string) []rdf.Triple {
	matchingTriples := make([]rdf.Triple, 0)
	for _, triple := range triples {
		match := true
		if subject != "" && triple.Subj.String() != subject {
			match = false
		}
		if predicate != "" && triple.Pred.String() != predicate {
			match = false
		}
		if object != "" && triple.Obj.String() != object {
			match = false
		}
		if match == true {
			matchingTriples = append(matchingTriples, triple)
		}
	}
	return matchingTriples
}

func writeTriplesToDisk(triples []rdf.Triple, filePath string, format rdf.Format) error {
	// ### Create encoder and add namespaces
	skosFileWriter, err := os.OpenFile(filePath, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModePerm)
	defer skosFileWriter.Close()
	if err != nil {
		return err
	}
	encoder := rdf.NewTripleEncoder(skosFileWriter, format)

	// # Overwrite working copy with revised triples
	encoder.EncodeAll(triples)
	err = encoder.Close()
	return err
}
