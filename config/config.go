package config

// How the public key is stored
// 0: external file, 1: external storage device, 2: tpm2, 3: usb serial
var KeyStore = 3

// Where the public key is stored, only applies for 0, 1 and 3
var KeyLocation = "/dev/ttyACM1"
