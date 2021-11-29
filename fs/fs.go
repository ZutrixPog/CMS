package fs

import(
	"io"
	"io/ioutil"
	"os"
)

func ReadMarkdown(filename string) (string, error) {
	file, err := os.Open("data/markdown/" + filename + ".md")
	if err != nil{
		return "", err
	}
	defer file.Close()

	markdown, err := ioutil.ReadAll(file)
	if err != nil{
		return "", err
	}
	return string(markdown), nil
}

func WriteMarkdown(filename string, markdown io.Reader) error {
	var file *os.File
	var err error
	if _, err := os.Stat("../data/markdown/" + filename + ".txt"); os.IsNotExist(err) {
		file, err = os.Create("../data/markdown/" + filename + ".txt")
	} else {
		file, err = os.Open("../data/markdown/" + filename + ".txt")
	}
	if err != nil{
		return err
	}
	defer file.Close()

	io.Copy(file, markdown)
	return nil
}

func RemoveMarkdown(filename string) error {
	err := os.Remove("data/markdown/" + filename + ".md")
	if err != nil {
		return err
	}
	return nil
}