package cmd

type Command struct {
	Name    string
	Command string
}
type Config struct {
	TmpPath        string
	Timeout        int
	Bucket         string
	ConnectionName string
	Commands       []Command
	SubPath        string
}
