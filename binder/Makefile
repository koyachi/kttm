logo:
	curl -O http://whatthingsdo.com/wp-content/uploads/2012/11/KeepingTwo1B1.png

index_html:
	go run binder.go > index.html

mobi: index_html
	kindlegen keeping_two.opf
