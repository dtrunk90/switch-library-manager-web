package web

import (
	"github.com/pierrre/imageserver"
	imageserver_http "github.com/pierrre/imageserver/http"
	imageserver_http_gift "github.com/pierrre/imageserver/http/gift"
	imageserver_image "github.com/pierrre/imageserver/image"
	_ "github.com/pierrre/imageserver/image/gif"
	imageserver_image_gift "github.com/pierrre/imageserver/image/gift"
	imageserver_source "github.com/pierrre/imageserver/source"
	_ "github.com/pierrre/imageserver/image/jpeg"
	_ "github.com/pierrre/imageserver/image/png"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

func (web *Web) HandleImages() {
	http.Handle("/i/", &imageserver_http.Handler {
		Parser: imageserver_http.ListParser([]imageserver_http.Parser {
			&imageserver_http.SourceTransformParser {
				Parser: &imageserver_http.SourcePathParser{},
				Transform: func(source string) string {
					return strings.TrimPrefix(source, "/i/")
				},
			},
			&imageserver_http_gift.ResizeParser{},
		}),
		Server: &imageserver.HandlerServer {
			Server: imageserver.Server(imageserver.ServerFunc(func(params imageserver.Params) (*imageserver.Image, error) {
				source, err := params.GetString(imageserver_source.Param)
				if err != nil {
					return nil, err
				}
				im, err := web.GetImg(source)
				if err != nil {
					return nil, &imageserver.ParamError{Param: imageserver_source.Param, Message: err.Error()}
				}
				return im, nil
			})),
			Handler: &imageserver_image.Handler {
				Processor: &imageserver_image_gift.ResizeProcessor{},
			},
		},
	})
}

func (web *Web) GetImg(name string) (*imageserver.Image, error) {
	filePath := filepath.Join(web.dataFolder, "img", name)
	data, err := ioutil.ReadFile(filePath)

	if err != nil {
		return nil, err
	}

	im := &imageserver.Image {
		Format: "jpeg",
		Data:   data,
	}

	return im, nil
}
