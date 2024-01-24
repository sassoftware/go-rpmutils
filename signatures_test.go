/*
 * Copyright (c) SAS Institute Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpmutils

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
)

func TestSign(t *testing.T) {
	keyring, err := openpgp.ReadArmoredKeyRing(bytes.NewReader([]byte(testkey)))
	if err != nil {
		t.Fatal("failed to parse test key:", err)
	}
	entity := keyring[0]

	f, err := os.Open("testdata/simple-1.0.1-1.i386.rpm")
	if err != nil {
		t.Fatal("failed to open test rpm:", err)
	}
	defer f.Close()
	h, err := SignRpmStream(f, entity.PrivateKey, nil)
	if err != nil {
		t.Fatal("error signing rpm:", err)
	}
	sigblob, err := h.DumpSignatureHeader(false)
	if err != nil {
		t.Fatal("error writing sig header:", err)
	}
	if len(sigblob)%8 != 0 {
		t.Fatalf("incorrect padding: got %d bytes, expected a multiple of 8", len(sigblob))
	}
	// verify by merging the new sig header with the original file
	if _, err = f.Seek(int64(h.OriginalSignatureHeaderSize()), io.SeekStart); err != nil {
		t.Fatal("error seeking:", err)
	}
	signed := io.MultiReader(bytes.NewReader(sigblob), f)
	_, sigs, err := Verify(signed, keyring)
	if err != nil {
		t.Fatal("error verifying signature:", err)
	}
	if len(sigs) != 2 || sigs[0].Signer != entity || sigs[1].Signer != entity {
		t.Fatalf("error verifying signature: incorrect signers. found: %#v", sigs)
	}
	// check padding for odd sized signature tags
	h.sigHeader.entries[1234] = entry{dataType: RPM_BIN_TYPE, count: 3, contents: []byte("foo")}
	sigblob, err = h.DumpSignatureHeader(false)
	if err != nil {
		t.Fatal("error writing sig header:", err)
	}
	if len(sigblob)%8 != 0 {
		t.Fatalf("incorrect padding: got %d bytes, expected a multiple of 8", len(sigblob))
	}
}

const testkey = `
-----BEGIN PGP PRIVATE KEY BLOCK-----

lQOYBFnCxAYBCACsNEYGCSm1uh9PDiB2L7TudKRqrLBiArETQagzbyNuSHHNibLa
85u9X1ZqcPspqISQjTmk7zYCFWzXlMDzPvAeLqiLX/0NHsqMuFFCGSE5jH0uS+KN
P4eLBYYgAJFCa4foyIGpESg52GA2/wZfvF5NOen0irh9XaA869jcWjb3c1euKLKo
0DEJU6OoHeiAo9SHJOicVddVUz+pigJ/++4bCwPxTH6ohx72ZwTknCXifjeqcasI
t76eTgTzwdSaOKLB1HWasauH5R7AW1oCgvqGBXNRKq1aR85avEVrUEEAyymk9Moy
9Hfm8XfZ2zsEMlJUsw9/F/oO3vqLkzSCCcTXABEBAAEAB/wN/5vXnsQQvUYRR500
7lDfd4TsFQirlvttDM/PCpBPRT1XD4QGD3qQDOF5+qA4NTY9h/VxJm72AWbdKX77
5xhe470Yw19PQzsE8HDOljtgsb51Vn7eq5TppLPQAyvLwfEE59O+eiISfbfokJek
jav+zB/sHKC9tDAz85on43+HYutLTS53AfJMdhzCMxpt2jwEyGPH0Ti+4yAeOsSI
v+J8YMHYeqMMp5Z1uWBEo4Kdh3R5BMNg2ovmW311ZW3dK363TG84jnhumU0yAaKy
DOsLy6xM4Sm617JQn4oe4YWgfjcmAsFo5Ek78UHqnHA6qJtHmQqUtGJFXPhvR2Mq
0tstBADKReEkQiTsIYoQvJmmu6ShiNVJ2KtSezkE/ZtK+Ne9ww+5upAwWkB+FOxM
+m53LkuKe8wK8ucIPb2ybVL3bqQb1REFbhf6o1H5mYnMKizcL3p0THuabtG9BG2Y
/wt+hNw9nAhPuS8yQ7tETYHGPfdl7221qxhDO5QDlDBRmqXHKwQA2fHId+1po4BV
ovRdbJxJ2uNhx+93RJORR3XnIs3tOrwD7bmt/B8zqoxi3FZ/414bwV2VPo6TMWV4
bNC6S0D+j3z2QLkGVp9woRaiC1+ZULwjugMl4Ou6oZNXT69wcGjdLw6rrvEl09y0
/qw3GzMgCn2ePVI16yqwV18wN662IwUD/1WvLpIyoCSALdp2lc17we+qbz/3Js/g
tfkkBj/xP8GVZd+xnFHHoQ6EO8RFTstC6mCIDMKjkvaPJmqxOLdJeK1gpRIjIoj1
o6JvpEfapy/xb/XV9EVikmIjt+wNY9V1JkU0u8o85uirHdzi3atXd8EVR5u/Zejb
ll2lNE7o1ltLRIu0CHRlc3Qga2V5iQFOBBMBCAA4FiEEttzyqc+V4Go/Bkjc3/2e
kI1I8igFAlnCxAYCGwMFCwkIBwIGFQgJCgsCBBYCAwECHgECF4AACgkQ3/2ekI1I
8ijrmwf6A1Bixs6NwT/LPW3MqjHW5n6FmoiZXBzNnOeBHk6FPI1qAADeZAQPMTq3
gKG2J5ciBQhpKGGqT31ovKkhlnpKaGUIaj8IAA7rI5UlbOTfTqVmjtpfYm43IGdl
gccZvlxtWWKGYZSyMHg2DEC6SJYpR9AHxbh4UvKFuTx9hnpWjVasOqqIl0Zs+fT4
W5FHS9C5kxrA67+9Wn7V8RY0aXn0zPvg8KUzmGMeovt7bYRvK+l58MVMupQ/m01S
pGgCzr9O7MAYsuJiWG7QoNriR8QsbAfsD70eNFSk4xKbpqXCqARfnHkDBU95WC57
bCw9mwgJ2r0mQLqjrXjEYBhaE49I8A==
=+d52
-----END PGP PRIVATE KEY BLOCK-----
`
