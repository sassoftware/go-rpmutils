package rpmutils

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestUncompress(t *testing.T) {
	defer goleak.VerifyNone(t)
	// built with e.g.: rpmbuild -bb payload-test.spec -D'_binary_payload w.ufdio'
	payloadTypes := []string{"w3.zstdio", "w6.lzdio", "w6.xzdio", "w9.bzdio", "w9.gzdio", "w.ufdio"}
	for _, payloadType := range payloadTypes {
		t.Run(payloadType, func(t *testing.T) {
			// open rpm
			fp := filepath.Join("testdata", "payload-test-0.1-"+payloadType+".x86_64.rpm")
			f, err := os.Open(fp)
			require.NoError(t, err)
			defer f.Close()
			rpm, err := ReadRpm(f)
			require.NoError(t, err)
			// consume payload
			var files int
			payload, err := rpm.PayloadReaderExtended()
			require.NoError(t, err)
			for {
				_, err := payload.Next()
				if err == io.EOF {
					break
				}
				require.NoError(t, err)
				_, err = io.Copy(io.Discard, payload)
				require.NoError(t, err)
				files++
			}
			assert.Equal(t, 1, files)
		})
	}
}

func TestUncompressEmpty(t *testing.T) {
	f, err := os.Open("testdata/empty-0.1-1.x86_64.rpm")
	require.NoError(t, err)
	defer f.Close()
	rpm, err := ReadRpm(f)
	require.NoError(t, err)
	payload, err := rpm.PayloadReaderExtended()
	require.NoError(t, err)
	_, err = payload.Next()
	assert.ErrorIs(t, err, io.EOF)
}
