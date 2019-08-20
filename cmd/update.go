package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/metal-pod/v"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cheggaaa/pb/v3"
)

const (
	downloadURLPrefix = "https://blobstore.fi-ts.io/metal/cloudctl/"
	releaseURL        = downloadURLPrefix + "version-" + runtime.GOOS + "-" + runtime.GOARCH + ".json"
	binaryURL         = downloadURLPrefix + programName + "-" + runtime.GOOS + "-" + runtime.GOARCH
)

// FIXME duplicate from metalctl, move to package on github ?
var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "update the program",
	}
	updateCheckCmd = &cobra.Command{
		Use:   "check",
		Short: "check for update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateCheck()
		},
	}
	updateDoCmd = &cobra.Command{
		Use:   "do",
		Short: "do the update of the program",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateDo()
		},
	}
	updateDumpCmd = &cobra.Command{
		Use:   "dump <binary>",
		Short: "dump the version update file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return updateDump(args)
		},
	}
)

// Release represents a release
type Release struct {
	Version  time.Time
	Checksum string
}

func init() {
	updateCmd.AddCommand(updateCheckCmd)
	updateCmd.AddCommand(updateDoCmd)
	updateCmd.AddCommand(updateDumpCmd)
}

func updateDo() error {
	latestVersion, err := getVersionInfo(releaseURL)
	if err != nil {
		return fmt.Errorf("unable read version information:%v", err)
	}

	tmpFile, err := ioutil.TempFile("", programName)
	if err != nil {
		return fmt.Errorf("unable create tempfile:%v", err)
	}
	err = downloadFile(tmpFile, binaryURL, latestVersion.Checksum)
	if err != nil {
		return err
	}
	location, err := getOwnLocation()
	if err != nil {
		return fmt.Errorf("unable to get own binary location:%v", err)
	}
	info, err := os.Stat(location)
	if err != nil {
		return fmt.Errorf("unable to stat old binary:%v", err)
	}
	mode := info.Mode()
	lf, err := os.OpenFile(location, os.O_WRONLY, mode)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("unable to write to:%s need root access:%v", location, err)
		}
	}
	lf.Close()

	oldlocation := location + ".update"
	defer os.Remove(oldlocation)
	err = os.Rename(location, oldlocation)
	if err != nil {
		return fmt.Errorf("unable to rename old binary:%v", err)
	}

	err = copy(tmpFile.Name(), location)
	if err != nil {
		return fmt.Errorf("unable to copy:%v", err)
	}
	err = os.Chmod(location, mode)
	if err != nil {
		return fmt.Errorf("unable to chown:%v", err)
	}

	return nil
}

func updateDump(args []string) error {
	if len(args) < 1 {
		return errors.New("binary argument required")
	}

	location := args[0]

	checksum, err := sum(location)
	if err != nil {
		return err
	}
	version, err := time.Parse(time.RFC3339, v.BuildDate)
	if err != nil {
		return err
	}
	r := Release{
		Version:  version,
		Checksum: checksum,
	}
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(bytes))
	return nil
}

func updateCheck() error {
	latestVersion, err := getVersionInfo(releaseURL)
	if err != nil {
		return err
	}
	version, err := time.Parse(time.RFC3339, v.BuildDate)
	if err != nil {
		return err
	}
	age := version.Sub(latestVersion.Version)
	location, err := getOwnLocation()
	if err != nil {
		return err
	}
	fmt.Printf("latest version:%s\n", latestVersion.Version.Format(time.RFC3339))
	fmt.Printf("local  version:%s\n", version.Format(time.RFC3339))

	if age > 24*time.Hour {
		fmt.Printf("%s is %s old, please run '%s update do'\n", programName, humanizeDuration(age), programName)
		fmt.Printf("%s location:%s\n", programName, location)
	} else {
		fmt.Printf("%s is up to date\n", programName)
	}
	return nil
}

func getOwnLocation() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	location, err := filepath.EvalSymlinks(ex)
	if err != nil {
		return "", err
	}
	return location, nil
}

func sum(binary string) (string, error) {
	hasher := sha256.New()
	s, err := ioutil.ReadFile(binary)
	hasher.Write(s)
	if err != nil {
		return "", err
	}

	return string(hex.EncodeToString(hasher.Sum(nil))), nil
}

func getVersionInfo(url string) (Release, error) {
	var release Release
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return release, errors.Wrap(err, "error creating new http request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return release, errors.Wrapf(err, "error with http GET for endpoint %s", url)
	}

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return release, errors.Wrap(err, "Error getting json from "+programName+" version url")
	}
	return release, nil
}

// downloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(out *os.File, url, checksum string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	defer out.Close()
	fileSize := resp.ContentLength

	bar := pb.Full.Start64(fileSize)
	// create proxy reader
	barReader := bar.NewProxyReader(resp.Body)
	_, err = io.Copy(out, barReader)
	bar.Finish()

	c, err := sum(out.Name())
	if err != nil {
		return fmt.Errorf("unable to calculate checksum:%v", err)
	}
	if c != checksum {
		return fmt.Errorf("checksum mismatch %s:%s", c, checksum)
	}

	return err
}

func copy(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dst, input, os.ModeType)
	if err != nil {
		return err
	}
	return nil
}
func humanizeDuration(duration time.Duration) string {
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
