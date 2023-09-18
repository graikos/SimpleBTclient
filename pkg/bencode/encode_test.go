package bencode

import "testing"

func TestEncodeBencodeToString(t *testing.T) {
	var res string
	var err error
	res, err = EncodeBencodeToString(15)
	if err != nil {
		t.Fatal(err)
	}
	if res != "i15e" {
		t.Fatalf("error encoding int")
	}

	res, err = EncodeBencodeToString("hello")
	if err != nil {
		t.Fatal(err)
	}
	if res != "5:hello" {
		t.Fatalf("error encoding string")
	}

	res, err = EncodeBencodeToString([]interface{}{"hello", 5})
	if err != nil {
		t.Fatal(err)
	}
	if res != "l5:helloi5ee" {
		t.Fatalf("error encoding simple list")
	}

	res, err = EncodeBencodeToString(map[string]interface{}{
		"info": []string{"ena", "tria"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res != "d4:infol3:ena4:triaee" {
		t.Fatalf("error encoding dictonary")
	}

}
