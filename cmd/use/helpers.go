package use

import (
	"fmt"
	"os"
	"io"
	"net/http"
	"archive/tar"
	"compress/gzip"
	"archive/zip"
	"path/filepath"
)

func GetHeaderFile(path string) (string,error) {
	file, err := os.Open(path)
    if (err != nil) {
		fmt.Fprintln(os.Stderr,"error: cannot open "+path+": "+err.Error())
		return "", err
	}
	defer file.Close()

	buff := make([]byte, 512)

	_, err = file.Read(buff)

    if err != nil {
        fmt.Fprintln(os.Stderr,"error reading file header: "+err.Error())
        return "", err
    }
	filetype := http.DetectContentType(buff)

	return filetype, nil
}

func isTarFile(path string) (bool,error) {
	file, err := os.Open(path)
    if (err != nil) {
		fmt.Fprintln(os.Stderr,"error: cannot open "+path+": "+err.Error())
		return false, err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	_, err = tarReader.Next()
    if (err != nil) {
        return false, nil
	}

    return true,nil
}

func isZip(path string) (bool,error) {
    header, err := GetHeaderFile(path)
	if (err == nil ) {
		return header == "application/zip", nil
	}
	return false, err
}

func isGzip(path string) (bool,error) {
    header, err := GetHeaderFile(path)
	if (err == nil ) {
		return header == "application/x-gzip", nil
	}
	return false, err
}

func IsCompressedFile(path string) (bool,error) {
	result, err := isGzip(path)
	if (err !=nil) {
	   return false,err
	} else if (result == true) {
		return result, nil
	}
	result, err = isZip(path)
	if (err !=nil) {
		return false,err
	 } else if (result == true) {
		 return result, nil
	 }
	 result, err = isTarFile(path)
	 if (err !=nil) {
		return false,err
	 }
	 return result,nil
}


func CopyFile(path string,destinationfile string) error {
	source, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr,"error opening file "+path+": "+err.Error())
		return  err
	}
	defer source.Close()
	dest, err := os.Create(destinationfile)
	if err != nil {
		fmt.Fprintln(os.Stderr,"error creating file "+destinationfile+": "+err.Error())
		return err
	}
    defer dest.Close()
	_, err = io.Copy(dest, source)
	if (err != nil) {
		fmt.Fprintln(os.Stderr,"error copying file "+path+" to "+destinationfile+": "+err.Error())
	}
	return err
}

func DecompressFile(path string,outpath string) (string,error) {
	fmt.Println("decompressing file "+path+" in "+outpath)
    var mgRootDir string = "" 
	result, err := isGzip(path)
	if ( err == nil ) {
	    if ( result ) {
           mgRootDir,err = ExtractTarGz(path,outpath)
		} else {
			result, err := isTarFile(path)
			if ( err == nil ) {
				if (result) {
					mgRootDir,err = ExtractTar(path,outpath)
				} else {
					result, err := isZip(path)
					if ( err == nil ) {
						if (result) {
						     mgRootDir,err = ExtractZip(path,outpath)
						}
					}
				}
			}
		}
	}

	return mgRootDir,err
}


func ExtractTarStream(st io.Reader,destinationdir string) (string,error) {
	firstDirectory := false
	var mgRootDir string = ""
    tarReader := tar.NewReader(st)

    for true {
        header, err := tarReader.Next()

        if err == io.EOF {
            break
        }

        if err != nil {
            fmt.Fprintln(os.Stderr,"cannot extract tar: " + err.Error())
			return "",err
        }

        switch header.Typeflag {
        case tar.TypeDir:
			if (!firstDirectory) {
				firstDirectory = true
				mgRootDir = destinationdir+"/"+header.Name
			}
            if err := os.Mkdir(destinationdir+"/"+header.Name, 0755); err != nil {
				fmt.Fprintln(os.Stderr,"mkdir failed extracting tar: "+err.Error())
				return "",err
            }
        case tar.TypeReg:
            outFile, err := os.Create(destinationdir+"/"+header.Name)
            if err != nil {
				fmt.Fprintln(os.Stderr,"create file failed extracting tar: "+err.Error())
				return "",err
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
				fmt.Fprintln(os.Stderr,"copy file failed extracting tar: "+err.Error())
				return "",err
            }
            outFile.Close()

        default:
			fmt.Fprintf(os.Stderr,"unknown type(%s) in %s: "+err.Error(),header.Typeflag,header.Name)
			return "",err
        }

    }
	return mgRootDir,nil
}

func ExtractTar(tarfile string,destinationdir string) (string,error) {
	tarStream, err := os.Open(tarfile)
	var mgRootDir string;
    if (err != nil) {
		fmt.Fprintln(os.Stderr,"error: cannot open "+tarfile+": "+err.Error())
		return "",err
	}
	defer tarStream.Close()

	var fileReader io.ReadCloser = tarStream
	mgRootDir, err = ExtractTarStream(fileReader,destinationdir)

	return mgRootDir, err
}

func ExtractZip(zipfile string,destinationdir string) (string,error) {

	firstDirectory := false
	var mgRootDir string = ""
	archive, err := zip.OpenReader(zipfile)
    if err != nil {
		fmt.Fprintln(os.Stderr,"error: cannot uncompress zip "+zipfile+": "+err.Error())
		return "",err
    }
	defer archive.Close()

    for _, f := range archive.File {
        filePath := filepath.Join(destinationdir, f.Name)

        if f.FileInfo().IsDir() {
			if (!firstDirectory) {
				firstDirectory = true
				mgRootDir = filePath
			}
            err = os.MkdirAll(filePath, os.ModePerm)
			if (err != nil) {
				fmt.Fprintln(os.Stderr,"error: cannot create directory "+filePath+": "+err.Error())
				return "",err
			}
        } else { 
            dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
            if err != nil {
				fmt.Fprintln(os.Stderr,"error: cannot create file "+filePath+": "+err.Error())
                return "",err
            }
			defer dstFile.Close()


            fileInArchive, err := f.Open()
            if err != nil {
				fmt.Fprintln(os.Stderr,"error: cannot open file "+f.Name+": "+err.Error())
                return "",err
            }
			defer fileInArchive.Close()

            if _, err := io.Copy(dstFile, fileInArchive); err != nil {
				fmt.Fprintln(os.Stderr,"error: cannot copy file to "+dstFile.Name()+": "+err.Error())
                return "",err
            }
		}
    }

	return mgRootDir,err
}

func ExtractTarGz(gzipfile string,destinationdir string) (string,error) {
	gzipStream, err := os.Open(gzipfile)
    if (err != nil) {
		fmt.Fprintln(os.Stderr,"error: cannot open "+gzipfile+": "+err.Error())
		return "",err
	}
	defer gzipStream.Close()
    uncompressedStream, err := gzip.NewReader(gzipStream)
    if err != nil {
		fmt.Fprintln(os.Stderr,"error: cannot uncompress gzip "+gzipfile+": "+err.Error())
		return "",err
    }
	return ExtractTarStream(uncompressedStream,destinationdir) 
}
