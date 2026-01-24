package cli

import (
	"fmt"
	"sort"
)

type branchStatusInfo struct {
	local         bool
	remoteMatches []string
}

func branchStatus(root string, branch string) (branchStatusInfo, error) {
	if branch == "" {
		return branchStatusInfo{}, fmt.Errorf("empty branch")
	}
	local, err := refExists(root, "refs/heads/"+branch)
	if err != nil {
		return branchStatusInfo{}, err
	}
	if local {
		return branchStatusInfo{local: true}, nil
	}

	remotes, err := gitRemotes(root)
	if err != nil {
		return branchStatusInfo{}, err
	}
	status := branchStatusInfo{remoteMatches: nil}
	for _, remote := range remotes {
		ok, err := refExists(root, "refs/remotes/"+remote+"/"+branch)
		if err != nil {
			return branchStatusInfo{}, err
		}
		if ok {
			status.remoteMatches = append(status.remoteMatches, remote)
		}
	}
	sort.Strings(status.remoteMatches)
	return status, nil
}
