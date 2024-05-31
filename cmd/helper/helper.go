package helper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"
)

// HumanizeDuration format given duration human readable
func HumanizeDuration(duration time.Duration) string {
	days := int64(duration.Hours() / 24)
	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))
	seconds := int64(math.Mod(duration.Seconds(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"d", days},
		{"h", hours},
		{"m", minutes},
		{"s", seconds},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		default:
			parts = append(parts, fmt.Sprintf("%d%s", chunk.amount, chunk.singularName))
		}
	}

	if len(parts) == 0 {
		return "0s"
	}
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}

func HumanizeSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// Prompt the user to given compare text
func Prompt(msg, compare string) error {
	fmt.Print(msg + " ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	text := scanner.Text()
	if text != compare {
		return fmt.Errorf("unexpected answer given (%q), aborting...", text)
	}
	return nil
}

// Truncate will trim a string in the middle and replace it with ellipsis
// FIXME write a test
func Truncate(input, ellipsis string, maxlength int) string {
	il := len(input)
	el := len(ellipsis)
	if il <= maxlength {
		return input
	}
	if maxlength <= el {
		return input[:maxlength]
	}
	startlength := ((maxlength - el) / 2) - el/2

	output := input[:startlength] + ellipsis
	missing := maxlength - len(output)
	output = output + input[il-missing:]
	return output
}

// ReadFrom will either read from stdin (-) or a file path an marshall from yaml to data
func ReadFrom(from string, data interface{}, f func(target interface{})) error {
	var reader io.Reader
	var err error
	switch from {
	case "-":
		reader = os.Stdin
	default:
		reader, err = os.Open(from)
		if err != nil {
			return fmt.Errorf("unable to open %s %w", from, err)
		}
	}
	dec := yaml.NewDecoder(reader)
	for {
		err := dec.Decode(data)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("decode error: %w", err)
		}
		f(data)
	}
	return nil
}

// Edit a yaml response from getFunc in place and call updateFunc after save
func Edit(id string, getFunc func(id string) ([]byte, error), updateFunc func(filename string) error) error {
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		editor = "vi"
	}

	tmpfile, err := os.CreateTemp("", "cloudctl*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())
	content, err := getFunc(id)
	if err != nil {
		return err
	}
	err = os.WriteFile(tmpfile.Name(), content, os.ModePerm) //nolint:gosec
	if err != nil {
		return err
	}
	editCommand := exec.Command(editor, tmpfile.Name())
	editCommand.Stdout = os.Stdout
	editCommand.Stdin = os.Stdin
	editCommand.Stderr = os.Stderr
	err = editCommand.Run()
	if err != nil {
		return err
	}
	return updateFunc(tmpfile.Name())
}

func LabelsToMap(labels []string) (map[string]string, error) {
	labelMap := make(map[string]string)
	for _, l := range labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("provided labels must be in the form <key>=<value>, found: %s", l)
		}
		labelMap[parts[0]] = parts[1]
	}
	return labelMap, nil
}

func MustPrintKubernetesResource(in any) {
	y, err := k8syaml.Marshal(in)
	if err != nil {
		panic(fmt.Errorf("unable to marshal to yaml: %w", err))
	}
	fmt.Printf("---\n%s", string(y))
}

func ClientNoAuth() runtime.ClientAuthInfoWriterFunc {
	noAuth := func(_ runtime.ClientRequest, _ strfmt.Registry) error { return nil }
	return runtime.ClientAuthInfoWriterFunc(noAuth)
}

func ExpandHomeDir(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to expand home dir: %w", err)
	}

	return filepath.Join(homedir, strings.TrimLeft(path, "~/")), nil
}
