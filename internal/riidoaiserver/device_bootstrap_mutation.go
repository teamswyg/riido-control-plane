package riidoaiserver

import (
	"math"
	"reflect"
)

func (m deviceBootstrapMutation) valid() error {
	a, b := m.after, m.before
	if a.valid() != nil {
		return errCLIInstallationInvalid
	}
	if b == nil {
		if a.State != "issued" || a.Revision != 1 || a.PolicyRevision != 0 {
			return errCLIInstallationConflict
		}
		return nil
	}
	if b.valid() != nil || b.Revision == math.MaxUint64 || a.Revision != b.Revision+1 || a.Request != b.Request || !reflect.DeepEqual(a.Before, b.Before) || !reflect.DeepEqual(a.After, b.After) {
		return errCLIInstallationConflict
	}
	switch {
	case (b.State == "issued" || b.State == "prepared") && a.State == b.State:
		if a.Intent.ID == b.Intent.ID || a.PolicyRevision != b.PolicyRevision || a.PreviousBootDigest != b.PreviousBootDigest {
			return errCLIInstallationConflict
		}
	case b.State == "issued" && a.State == "prepared":
		if a.Intent != b.Intent || a.PolicyRevision == 0 || a.PolicyRevision == math.MaxUint64 {
			return errCLIInstallationConflict
		}
	case b.State == "prepared" && a.State == "complete":
		if a.Intent != b.Intent || a.PreviousBootDigest != b.PreviousBootDigest || b.PolicyRevision == math.MaxUint64 || a.PolicyRevision != b.PolicyRevision+1 {
			return errCLIInstallationConflict
		}
	default:
		return errCLIInstallationConflict
	}
	return nil
}
