package endorse

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSaveAndLoad(t *testing.T) {
	testData := []struct {
		Name        string
		Password    []byte
		AuthResults []AuthenticationResult
	}{
		{
			Name:     "pattern 1",
			Password: []byte("password"),
			AuthResults: []AuthenticationResult{
				{
					KeyID: "key1",
					IMSI:  "11111",
					CK:    []byte{0x01, 0x02, 0x03, 0x04},
				},
			},
		},
	}

	for _, data := range testData {
		data := data // capture
		t.Run(data.Name, func(t *testing.T) {
			t.Parallel()

			kc1 := newKeyCacheImpl()
			for _, ar := range data.AuthResults {
				kc1.saveAuthResult(ar.IMSI, &ar)
			}

			buf := bytes.NewBuffer([]byte{})
			err := kc1.save(buf, data.Password)
			if err != nil {
				t.Fatalf("error occurred while calling save(): %+v", err)
			}

			kc2 := newKeyCacheImpl()
			err = kc2.load(buf, data.Password)
			if err != nil {
				t.Fatalf("error occurred while calling load(): %+v", err)
			}

			if !cmp.Equal(kc1, kc2, cmp.AllowUnexported(keyCacheImpl{}, keyCacheEntry{})) {
				t.Fatalf("\nExpected: %+v\nActual:   %+v\n", kc1, kc2)
			}
		})
	}

}

func TestEncodeAndDecode(t *testing.T) {
	testData := []struct {
		Name        string
		IV          []byte
		Key         []byte
		PlainData   []byte
		EncodedData []byte
	}{
		{
			Name: "pattern 1",
			IV: []byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
			},
			Key:       []byte("testP@$$w0rd"),
			PlainData: []byte("hello"),
			EncodedData: []byte{
				0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
				0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
				0x70, 0x26, 0xd5, 0x66, 0xc6,
			},
		},
	}

	for _, data := range testData {
		data := data // capture
		t.Run(data.Name, func(t *testing.T) {
			t.Parallel()

			encodedData, err := encode(data.PlainData, data.Key, data.IV)
			if err != nil {
				t.Fatalf("error occurred while calling encode(): %+v", err)
			}

			if !reflect.DeepEqual(data.EncodedData, encodedData) {
				t.Fatalf("\nExpected: %+v\nActual:   %+v\n", data.EncodedData, encodedData)
			}

			plainData, err := decode(encodedData, data.Key)
			if err != nil {
				t.Fatalf("error occurred while calling decode(): %+v", err)
			}
			if !reflect.DeepEqual(data.PlainData, plainData) {
				t.Fatalf("\nExpected: %+v\nActual:   %+v\n", data.PlainData, plainData)
			}
		})
	}
}

func TestMakeKeyFromPassword(t *testing.T) {
	testData := []struct {
		Name     string
		Password []byte
		Key      []byte
	}{
		{
			Name:     "pattern 1",
			Password: []byte("testP@$$w0rd"),
			Key:      []byte("testP@$$w0rd\x00\x00\x00\x00"),
		},
		{
			Name:     "pattern 2",
			Password: []byte("0123456789abcdef"),
			Key:      []byte("0123456789abcdef"),
		},
		{
			Name:     "pattern 3",
			Password: []byte("very very very very very long password"),
			Key:      []byte("very very very v"),
		},
	}

	for _, data := range testData {
		data := data // capture
		t.Run(data.Name, func(t *testing.T) {
			t.Parallel()

			key := makeKeyFromPassword(data.Password)
			if !reflect.DeepEqual(data.Key, key) {
				t.Fatalf("\nExpected: %+v\nActual:   %+v\n", data.Key, key)
			}

		})
	}
}
