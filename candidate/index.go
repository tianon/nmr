package candidate

import (
	"bufio"
	"fmt"
	"io"

	"pault.ag/go/debian/control"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/debian/version"
)

type Canidates map[string][]control.BinaryIndex

func (can *Canidates) AppendBinaryIndexReader(in io.Reader) error {
	reader := bufio.NewReader(in)
	index, err := control.ParseBinaryIndex(reader)
	if err != nil {
		return err
	}
	can.AppendBinaryIndex(index)
	return nil
}

func (can *Canidates) AppendBinaryIndex(index []control.BinaryIndex) {
	for _, entry := range index {
		(*can)[entry.Package] = append((*can)[entry.Package], entry)
	}
}

func NewCanidates(index []control.BinaryIndex) Canidates {
	ret := Canidates{}
	ret.AppendBinaryIndex(index)
	return ret
}

func ReadFromBinaryIndex(in io.Reader) (*Canidates, error) {
	reader := bufio.NewReader(in)
	index, err := control.ParseBinaryIndex(reader)
	if err != nil {
		return nil, err
	}
	can := NewCanidates(index)
	return &can, nil
}

func (can Canidates) ExplainSatisfiesBuildDepends(arch dependency.Arch, depends dependency.Dependency) (bool, string) {
	for _, possi := range depends.GetPossibilities(arch) {
		can, why := can.ExplainSatisfies(possi)
		if !can {
			return false, fmt.Sprintf("Possi %s can't be satisfied - %s", possi.Name, why)
		}
	}
	return true, "All relations are a go"
}

func (can Canidates) SatisfiesBuildDepends(arch dependency.Arch, depends dependency.Dependency) bool {
	ret, _ := can.ExplainSatisfiesBuildDepends(arch, depends)
	return ret
}

func (can Canidates) Satisfies(possi dependency.Possibility) bool {
	ret, _ := can.ExplainSatisfies(possi)
	return ret
}

func (can Canidates) ExplainSatisfies(possi dependency.Possibility) (bool, string) {
	///
	///  XXX: DON'T IGNORE ARCHES
	///

	entries, ok := can[possi.Name]
	if !ok { // no known entries in the Index
		return false, fmt.Sprintf("Totally unknown package: %s", possi.Name)
	}

	if possi.Version == nil {
		return true, "Relation exists, no version constraint"
	}

	// OK, so we have to play with versions now.
	vr := *possi.Version
	relatioNumber, _ := version.Parse(vr.Number)

	for _, installable := range entries {
		q := version.Compare(installable.Version, relatioNumber)

		explainMessage := fmt.Sprintf(
			"%s %%s %s (I only see %s)",
			possi.Name,
			vr.Number,
			installable.Version,
		)

		switch vr.Operator {
		case ">=":
			return q >= 0, fmt.Sprintf(explainMessage, ">=")
		case "<=":
			return q <= 0, fmt.Sprintf(explainMessage, "<=")
		case ">>":
			return q > 0, fmt.Sprintf(explainMessage, ">>")
		case "<<":
			return q < 0, fmt.Sprintf(explainMessage, "<<")
		case "=":
			return q == 0, fmt.Sprintf(explainMessage, "=")
		default:
			return false, "Unknown operator D:" // XXX: WHAT THE SHIT
		}
	}

	return false, "Unkown state"
}
