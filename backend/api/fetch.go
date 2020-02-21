package api

import (
	"io/ioutil"
	"net/http"

	"github.com/guad/paperless2/backend/crypto"

	"github.com/globalsign/mgo/bson"
	"github.com/guad/paperless2/backend/db"
	"github.com/guad/paperless2/backend/model"
	"github.com/guad/paperless2/backend/storage"
	"github.com/minio/minio-go"

	"github.com/labstack/echo"
)

func FetchFile(c echo.Context) error {
	doc := c.Param("doc")

	sesh := db.Ctx()
	defer sesh.Close()
	col := sesh.DB("paperless").C("documents")

	if !bson.IsObjectIdHex(doc) {
		return c.JSON(http.StatusBadRequest, struct{}{})
	}

	id := bson.ObjectIdHex(doc)

	var document model.Document

	err := col.FindId(id).One(&document)

	if err != nil {
		return err
	}

	key := document.S3Path

	file, err := storage.S3.GetObject(
		storage.DocumentBucket,
		key,
		minio.GetObjectOptions{},
	)

	if err != nil {
		return err
	}

	defer file.Close()

	// Decrypt
	encryptedBytes, err := ioutil.ReadAll(file)

	if err != nil {
		return err
	}

	decryptedBytes := crypto.Decrypt(encryptedBytes)

	return c.Blob(http.StatusOK, document.ContentType, decryptedBytes)
}
