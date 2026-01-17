package guard

import (
	"reflect"
	"testing"
)

func TestMergeTags(t *testing.T) {
	tests := []struct {
		name         string
		existingTags []string
		newTag       string
		want         []string
	}{
		{
			name:         "空のタグリストに新規タグを追加",
			existingTags: []string{},
			newTag:       "repo-a",
			want:         []string{"repo-a"},
		},
		{
			name:         "既存タグがある場合に新規タグを追加",
			existingTags: []string{"repo-a", "tag1"},
			newTag:       "repo-b",
			want:         []string{"repo-a", "tag1", "repo-b"},
		},
		{
			name:         "既に同じタグが存在する場合は追加しない",
			existingTags: []string{"repo-a", "tag1"},
			newTag:       "repo-a",
			want:         []string{"repo-a", "tag1"},
		},
		{
			name:         "newTagが空文字列の場合は変更なし",
			existingTags: []string{"repo-a", "tag1"},
			newTag:       "",
			want:         []string{"repo-a", "tag1"},
		},
		{
			name:         "nilのタグリストに新規タグを追加",
			existingTags: nil,
			newTag:       "repo-a",
			want:         []string{"repo-a"},
		},
		{
			name:         "newTagが空でexistingTagsがnil",
			existingTags: nil,
			newTag:       "",
			want:         nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeTags(tt.existingTags, tt.newTag)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeTags() = %v, want %v", got, tt.want)
			}
		})
	}
}
