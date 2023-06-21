package main

import "DaruBot/cmd"

// TODO makefile over docker
// go build -ldflags "-X DaruBot/cmd.GitCommit=$(git rev-list --abbrev-commit -1 HEAD) -X DaruBot/cmd.Ver=$(cat ./VERSION)-dev -X DaruBot/cmd.BuildDate=$(date --rfc-3339=date)"

func main() {
	// TODO use uber life cycle package

	cmd.Run()
}
