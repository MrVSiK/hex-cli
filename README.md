# hex-cli

This is the companion CLI for [hex-storage](https://github.com/MrVSiK/select). It allows an user to upload files to Azure Blob Storage using **hex-api**.

## Commands
```
hex upload -f FILEPATH
```
The above command will upload the file at the given path.
```
hex login -e EMAIL
```
This can be used to authorise a user. It will prompt a user for their password (chosen during registration).