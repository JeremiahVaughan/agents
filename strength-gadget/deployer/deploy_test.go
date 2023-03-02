package main

import (
	"sync"
	"testing"
)

func Test_areDirectoryContentsDirty(t *testing.T) {
	type args struct {
		p                    string
		directoryHash        string
		existingSourceHashes SourceHashes
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Dirty",
			args: args{
				p:             "/test/me/over/register",
				directoryHash: "P5o5JJ0-P",
				existingSourceHashes: SourceHashes{
					Hashes: sync.Map{},
				},
			},
			want: true,
		},
		{
			name: "Dirty 2",
			args: args{
				p:             "/test/me/over/register",
				directoryHash: "P5o5JJ0-P",
				existingSourceHashes: SourceHashes{
					Hashes: *newSyncMap("login", "P5o5JJ0-P"),
				},
			},
			want: true,
		},
		{
			name: "Dirty 3",
			args: args{
				p:             "/test/me/over/register",
				directoryHash: "P5o5JJ0-P",
				existingSourceHashes: SourceHashes{
					Hashes: *newSyncMap("register", "JGwSCYhNf"),
				},
			},
			want: true,
		},
		{
			name: "Clean",
			args: args{
				p:             "/test/me/over/register",
				directoryHash: "P5o5JJ0-P",
				existingSourceHashes: SourceHashes{
					Hashes: *newSyncMap("register", "P5o5JJ0-P"),
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := areDirectoryContentsDirty(tt.args.p, tt.args.directoryHash, &tt.args.existingSourceHashes); got != tt.want {
				t.Errorf("areDirectoryContentsDirty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newSyncMap(k string, v string) *sync.Map {
	syncMap := sync.Map{}
	syncMap.Store(k, v)
	return &syncMap
}

func Test_getLowestDirectory(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Happy path",
			args: args{
				path: "/home/ward/bound/lowest/some_file.txt",
			},
			want: "lowest",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLowestDirectory(tt.args.path); got != tt.want {
				t.Errorf("getLowestDirectory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_confirmUniqueNameOfDeploymentDirectories(t *testing.T) {
	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Happy Path",
			args: args{
				paths: []string{
					"/home/of/the/enchiladas/deployment_config.json",
					"/home/of/the/burritos/deployment_config.json",
					"/home/of/the/watermelon/deployment_config.json",
				},
			},
			wantErr: false,
		},
		{
			name: "Unhappy Path",
			args: args{
				paths: []string{
					"/home/of/the/enchiladas/deployment_config.json",
					"/home/of/the/burritos/deployment_config.json",
					"/home/of/the/burritos/deployment_config.json",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := confirmUniqueNameOfDeploymentDirectories(tt.args.paths); (err != nil) != tt.wantErr {
				t.Errorf("confirmUniqueNameOfDeploymentDirectories() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
