package fonts

import "encoding/base64"

func SiYuanHeiTiTTF() []byte {
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(siyuanheiti)))
	_, err := base64.StdEncoding.Decode(dbuf, []byte(siyuanheiti))
	if err != nil {
		panic(err)
	}
	return dbuf
}

func SiYuanHeiTiOTFBold() []byte {
	dbuf := make([]byte, base64.StdEncoding.DecodedLen(len(siyuanheitibold)))
	_, err := base64.StdEncoding.Decode(dbuf, []byte(siyuanheitibold))
	if err != nil {
		panic(err)
	}
	return dbuf
}
