package setting

import (
	"testing"
)

func TestOwnersFileToRepoInfo1(t *testing.T) {
	o := OwnersFile{
		EnableAutoMerge:      true,
		DeleteAfterAutoMerge: true,
	}

	ok, info := o.ToRepoInfo()
	if !ok {
		t.Errorf("should be success to convert from OwnersFile")
		return
	}

	if !info.EnableAutoMerge {
		t.Errorf("ShouldMergeAutomatically: should be true")
		return
	}

	if !info.DeleteAfterAutoMerge {
		t.Errorf("ShouldDeleteMerged: should be true")
		return
	}
}

func TestOwnersFileToRepoInfo2(t *testing.T) {
	o := OwnersFile{
		EnableAutoMerge:      false,
		DeleteAfterAutoMerge: false,
	}

	ok, info := o.ToRepoInfo()
	if !ok {
		t.Errorf("should be success to convert from OwnersFile")
		return
	}

	if info.EnableAutoMerge {
		t.Errorf("ShouldMergeAutomatically: should be false")
		return
	}

	if info.DeleteAfterAutoMerge {
		t.Errorf("ShouldDeleteMerged: should be false")
		return
	}
}
