package main

import (
	"testing"

	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_getServerVersion(t *testing.T) {
	type server struct {
		major string
		minor string
	}
	tests := []struct {
		name   string
		server server
		min    int
		want   bool
		err    bool
	}{
		{
			name: "minimal not respected",
			server: server{
				major: "1",
				minor: "9",
			},
			min:  10,
			want: false,
			err:  false,
		},
		{
			name: "minimal respected",
			server: server{
				major: "1",
				minor: "11",
			},
			min:  10,
			want: true,
			err:  false,
		},
		{
			name: "version of server is unreadable",
			server: server{
				major: "aze",
				minor: "11",
			},
			err: true,
		},
	}

	for _, tt := range tests {
		client := fake.NewSimpleClientset()

		fakeDiscovery, ok := client.Discovery().(*fakediscovery.FakeDiscovery)
		if !ok {
			t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
		}

		fakeDiscovery.FakedServerVersion = &version.Info{
			Major: tt.server.major,
			Minor: tt.server.minor,
		}

		res, err := checkMinimalServerVersion(client, 10)
		if err != nil != tt.err {
			t.Errorf("Expected error: %v\n", tt.err)
		}
		if res != tt.want {
			t.Errorf("Expected %v, got %v\n", tt.want, res)
		}
	}

}
