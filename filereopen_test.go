package filereopen

import (
	"bufio"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestReopen(t *testing.T) {
	tmpdir := t.TempDir()
	file := tmpdir + "/test.log"
	filemv1 := tmpdir + "/test.log.1"
	filemv2 := tmpdir + "/test.log.2"
	filemv3 := tmpdir + "/test.log.3"

	f, err := OpenFileForAppend(file, 0600)
	require.NoError(t, err)
	require.NoError(t, f.SetInterval(time.Millisecond*101))
	f.SetErrorFunction(func(e error) { t.Logf("e: %s", e) })
	for i := 0; i < 4000; i++ {
		time.Sleep(time.Millisecond)
		w := []byte(fmt.Sprintf("%08d\n", i))
		n, err := f.Write(w)
		require.NoError(t, err, "iter: %d", i)
		require.EqualValues(t, len(w), n, "full write [%d]", i)
		if i == 1200 {
			err := os.Rename(file, filemv1)
			assert.NoError(t, err, "rename log file 1")
		}
		if i == 2400 {
			err := os.Rename(filemv1, filemv2)
			assert.NoError(t, err, "rename log file 2")
			err = os.Rename(file, filemv1)
			assert.NoError(t, err, "rename log file 1")
		}
		if i == 3600 {
			err := os.Rename(filemv2, filemv3)
			assert.NoError(t, err, "rename log file 3")
			err = os.Rename(filemv1, filemv2)
			assert.NoError(t, err, "rename log file 2")
			err = os.Rename(file, filemv1)
			assert.NoError(t, err, "rename log file 1")
		}
	}
	err = f.Close()
	assert.NoError(t, err)
	lineList := map[string]bool{}

	for _, f := range []string{file, filemv1, filemv2, filemv3} {
		file, err := os.Open(f)
		assert.NoError(t, err)
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			l := scanner.Text()
			if _, found := lineList[l]; found {
				t.Errorf("found duplicate line [%s]", l)
			} else {
				lineList[l] = true
			}
		}
	}
	fail := false
	for i := 0; i < 4000; i++ {
		l := fmt.Sprintf("%08d", i)
		if _, found := lineList[l]; !found {
			t.Logf("does not contain line [%s]", l)
			fail = true
		}
	}
	if fail {
		t.FailNow()
	}
}
