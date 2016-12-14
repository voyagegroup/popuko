package main

import (
	"testing"
)

func TestRepositorySettingToRepoInfo1(t *testing.T) {
	o := RepositorySetting{
		shouldMergeAutomatically: true,
		shouldDeleteMerged:       true,
	}

	ok, info := o.ToRepoInfo()
	if !ok {
		t.Fatal("should be success to convert from OwnersFile")
	}

	if !info.ShouldMergeAutomatically {
		t.Fatal("ShouldMergeAutomatically: should be true")
	}

	if !info.DeleteAfterAutoMerge {
		t.Fatal("ShouldDeleteMerged: should be true")
	}
}
