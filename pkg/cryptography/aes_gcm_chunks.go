package cryptography

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"github.com/jaeha-choi/Proj_Coconut_Utility/log"
	"github.com/jaeha-choi/Proj_Coconut_Utility/util"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
)

const (
	// ChunkSize should be less than max value of uint32 (4294967295)
	// since the util package use unsigned 4 bytes to represent the data size.
	ChunkSize = 1.28e+8
	// MaxFileSize indicates theoretical limit for the file size. Because chunk number are
	// indicated with uint16, MaxFileSize depends on ChunkSize. However, actual file limit
	// may depend on the IV, file system, OS, etc.
	MaxFileSize = ChunkSize * 65535
	IvSize      = 12
	gcmOverhead = 16
)

var ChunkIncorrectOrder = errors.New("encrypted chunk incorrect order")

var IncompleteFile = errors.New("incomplete file")

type AesGcmChunk struct {
	key           []byte
	file          *os.File
	fileName      string
	readOffset    uint64
	readChunkNum  uint16
	writeOffset   uint64
	writeChunkNum uint16
	fileSize      uint64
	chunkCount    uint16
	isEncrypt     bool
}

func EncryptSetup(fileN string) (ag *AesGcmChunk, err error) {
	symKey, err := genSymKey()
	if err != nil {
		log.Debug(err)
		log.Error("Error while generating an AES key")
		return nil, err
	}
	srcFile, err := os.Open(fileN)
	if err != nil {
		log.Debug(err)
		log.Error("Error while opening a file")
		return nil, err
	}
	fileStat, err := os.Stat(fileN)
	if err != nil {
		log.Debug(err)
		log.Error("Error while getting stats")
		return nil, err
	}
	_, fileName := filepath.Split(fileN)
	fileSize := fileStat.Size()
	chunkNum := math.Ceil(float64(fileSize) / ChunkSize)
	return &AesGcmChunk{
		key:           symKey,
		file:          srcFile,
		fileName:      fileName,
		readOffset:    0,
		readChunkNum:  0,
		writeOffset:   0,
		writeChunkNum: 0,
		fileSize:      uint64(fileSize),
		chunkCount:    uint16(chunkNum),
		isEncrypt:     true,
	}, nil
}

func DecryptSetup() (ag *AesGcmChunk, err error) {
	// Create directory if it doesn't exist
	if err = os.MkdirAll(util.DownloadPath, os.ModePerm); err != nil {
		log.Debug(err)
		log.Error("Error while creating download directory")
		return nil, err
	}
	// Create temporary file for downloading
	tmpFile, err := ioutil.TempFile(util.DownloadPath, ".tmp_download_")
	if err != nil {
		log.Debug(err)
		log.Error("Temp file could not be created")
		return nil, err
	}
	return &AesGcmChunk{
		key:           nil,
		file:          tmpFile,
		fileName:      "",
		readOffset:    0,
		readChunkNum:  0,
		writeOffset:   0,
		writeChunkNum: 0,
		fileSize:      0,
		chunkCount:    0,
		isEncrypt:     false,
	}, nil
}

func (ag *AesGcmChunk) Encrypt(writer io.Writer, receiverPubKey *rsa.PublicKey, senderPrivKey *rsa.PrivateKey) (err error) {
	dataEncrypted, dataSignature, err := encryptSignSymKey(ag.key, receiverPubKey, senderPrivKey, util.Uint16ToByte(ag.chunkCount))
	if err != nil {
		log.Debug(err)
		log.Error("Error in encryptSignSymKey")
		return err
	}

	// Send dataEncrypted
	if _, err = util.WriteBytes(writer, dataEncrypted); err != nil {
		log.Debug(err)
		log.Error("Error in WriteBytes while sending dataEncrypted")
		return err
	}

	// Send dataSignature
	if _, err = util.WriteBytes(writer, dataSignature); err != nil {
		log.Debug(err)
		log.Error("Error in WriteBytes while sending dataSignature")
		return err
	}

	// Send file name
	encryptedFileName, iv, err := ag.encryptBytes([]byte(ag.fileName))
	if err != nil {
		log.Debug(err)
		log.Error("Error in encryptBytes while encrypting file name")
		return err
	}
	if _, err = util.WriteBytes(writer, iv); err != nil {
		log.Debug(err)
		log.Error("Error in WriteBytes while writing iv")
		return err
	}
	if _, err = util.WriteBytes(writer, encryptedFileName); err != nil {
		log.Debug(err)
		log.Error("Error in WriteBytes while writing encrypted file name")
		return err
	}

	// Send file
	var encryptedFileChunk, ivAndChNum []byte
	// ag.readOffset and ag.readChunkNum are updated in encryptChunk
	for ag.readOffset < ag.fileSize {
		if ag.readOffset+ChunkSize >= ag.fileSize {
			// Send last chunk
			encryptedFileChunk, ivAndChNum, err = ag.encryptChunk(ag.fileSize - ag.readOffset)
		} else {
			// Send chunk
			encryptedFileChunk, ivAndChNum, err = ag.encryptChunk(ChunkSize)
		}
		if err != nil {
			log.Debug(err)
			log.Error("Error in encryptChunk. Read Offset: ", int(ag.readOffset))
			return err
		}
		// Send ivAndChNum
		if _, err = util.WriteBytes(writer, ivAndChNum); err != nil {
			log.Debug(err)
			log.Error("Error in WriteBytes while sending ivAndChNum")
			return err
		}
		// Send encryptedFileChunk
		if _, err = util.WriteBytes(writer, encryptedFileChunk); err != nil {
			log.Debug(err)
			log.Error("Error in WriteBytes while sending encryptedFileChunk")
			return err
		}
	}
	return nil
}

func (ag *AesGcmChunk) encryptChunk(chunkSize uint64) (encryptedData []byte, ivAndChNum []byte, err error) {
	currChunkNum := util.Uint16ToByte(ag.readChunkNum)
	//plain := make([]byte, int(chunkSize)-gcmOverhead-len(currChunkNum))
	plain := make([]byte, int(chunkSize))
	if _, err := io.ReadFull(ag.file, plain); err != nil {
		log.Debug(err)
		log.Error("Error while reading src file")
		return nil, nil, err
	}

	encryptedData, iv, err := ag.encryptBytes(plain)
	if err != nil {
		log.Debug(err)
		log.Error("Error in encryptBytes")
		return nil, nil, err
	}

	ivAndChNum = append(iv, currChunkNum...)

	ag.readChunkNum += 1
	ag.readOffset += uint64(len(plain))
	//log.Debug(fmt.Sprintf("Len: %d, full: %s", len(res), BytesToBase64(res)))
	return encryptedData, ivAndChNum, err
}

func (ag *AesGcmChunk) encryptBytes(plain []byte) (encrypted []byte, iv []byte, err error) {
	block, err := aes.NewCipher(ag.key)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating new cipher block")
		return nil, nil, err
	}
	iv = make([]byte, IvSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Debug(err)
		log.Error("Error while creating iv")
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Debug(err)
		log.Error("Error in NewGCM")
		return nil, iv, err
	}
	// For concurrency/parallelism, additionalData could include file id
	res := aesgcm.Seal(nil, iv, plain, nil)
	return res, iv, nil
}

func (ag *AesGcmChunk) Decrypt(reader io.Reader, senderPubKey *rsa.PublicKey, receiverPrivKey *rsa.PrivateKey) (err error) {
	dataEncrypted, err := util.ReadBytes(reader)
	if err != nil {
		log.Debug(err)
		log.Error("Error in ReadBytes while getting dataEncrypted")
		return err
	}
	dataSignature, err := util.ReadBytes(reader)
	if err != nil {
		log.Debug(err)
		log.Error("Error in ReadBytes while getting dataEncrypted")
		return err
	}
	dataPlain, err := decryptVerifySymKey(dataEncrypted, dataSignature, senderPubKey, receiverPrivKey)
	if err != nil {
		log.Debug(err)
		log.Error("Error in decryptVerifySymKey")
		return err
	}
	ag.key = dataPlain[:SymKeySize]
	ag.chunkCount = util.ByteToUint16(dataPlain[SymKeySize:])

	ivFileName, err := util.ReadBytes(reader)
	if err != nil {
		log.Debug(err)
		log.Error("Error while reading iv for file name")
		return err
	}
	encryptedFileName, err := util.ReadBytes(reader)
	if err != nil {
		log.Debug(err)
		log.Error("Error while reading encrypted file name")
		return err
	}
	decryptedFileName, err := ag.decryptBytes(encryptedFileName, ivFileName)
	if err != nil {
		log.Debug(err)
		log.Error("Error while decrypting file name")
		return err
	}
	ag.fileName = string(decryptedFileName)

	var encryptedFileChunk, ivAndChNum []byte
	// ag.writeOffset and ag.writeChunkNum are updated in decryptChunk
	for ag.writeChunkNum < ag.chunkCount {
		if ivAndChNum, err = util.ReadBytes(reader); err != nil {
			log.Debug(err)
			log.Error("Error in ReadBytes while reading ivAndChNum")
			return err
		}
		if encryptedFileChunk, err = util.ReadBytes(reader); err != nil {
			log.Debug(err)
			log.Error("Error in ReadBytes while reading encryptedFileChunk")
			return err
		}
		decryptedFileChunk, _, err := ag.decryptChunk(encryptedFileChunk, ivAndChNum)
		if err != nil {
			log.Debug(err)
			log.Error("Error in decryptChunk")
			return err
		}
		if _, err = ag.file.Write(decryptedFileChunk); err != nil {
			log.Debug(err)
			log.Error("Error while writing decrypted file chunk to temp file")
			return err
		}
	}
	if ag.writeChunkNum != ag.chunkCount {
		log.Error("Not all chunks were received")
		return IncompleteFile
	}
	return nil
}

func (ag *AesGcmChunk) decryptChunk(encryptedData []byte, ivAndChNum []byte) (decryptedData []byte, currChunkNum uint16, err error) {
	iv := ivAndChNum[:IvSize]
	currChunkNum = util.ByteToUint16(ivAndChNum[IvSize:])
	decryptedData, err = ag.decryptBytes(encryptedData, iv)
	if err != nil {
		log.Debug(err)
		log.Error("Error in decryptBytes")
		return nil, 0, err
	}

	// If chunk was received in incorrect order, raise error
	if ag.writeChunkNum != currChunkNum {
		log.Error("Encrypted chunk was received in an incorrect order")
		return decryptedData, currChunkNum, ChunkIncorrectOrder
	}

	// Update variables
	ag.writeChunkNum += 1
	ag.writeOffset += uint64(len(decryptedData))
	return decryptedData, currChunkNum, nil
}

func (ag *AesGcmChunk) decryptBytes(encryptedData []byte, iv []byte) (decryptedData []byte, err error) {
	block, err := aes.NewCipher(ag.key)
	if err != nil {
		log.Debug(err)
		log.Error("Error while creating new cipher block")
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Debug(err)
		log.Error("Error in NewGCM")
		return nil, err
	}
	// Decrypt the encryptedData
	if decryptedData, err = aesgcm.Open(nil, iv, encryptedData, nil); err != nil {
		log.Debug(err)
		log.Error("Error in Open in decryptChunk")
		return nil, err
	}
	return decryptedData, nil
}

func (ag *AesGcmChunk) Close() (err error) {
	if err = ag.file.Close(); err != nil {
		log.Debug(err)
		log.Error("File failed to close")
		return err
	}
	// Only applies to decryption
	if !ag.isEncrypt {
		// If file was fully processed, rename temp file to actual name
		if ag.writeChunkNum == ag.chunkCount {
			// Move temporary file to download directory (DownloadPath)
			if err := os.Rename(ag.file.Name(), filepath.Join(util.DownloadPath, ag.fileName)); err != nil {
				log.Debug(err)
				log.Debug("Tmp file name: ", ag.file.Name())
				log.Debug("File name: ", ag.fileName)
				log.Error("Error moving the temp file to download path")
				// If rename was unsuccessful, remove file
				if err := os.Remove(ag.file.Name()); err != nil {
					log.Debug(err)
					log.Error("Error while removing temp file. Temp file at: ", ag.file.Name())
					return err
				}
				return err
			}
		} else {
			// If file was not fully processed, delete file
			if err := os.Remove(ag.file.Name()); err != nil {
				log.Debug(err)
				log.Error("Error while removing temp file. Temp file at: ", ag.file.Name())
				return err
			}
		}
	}
	return nil
}
