To install **Go (Golang) version 1.24.3** on **Ubuntu**, follow these steps:

1. **Download the Go 1.24.3 Binary**:
   - Open a terminal and download the Go 1.24.3 tarball using `wget`:
     ```bash
     wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
     ```
2. **Extract the Tarball**:

   - Extract the downloaded file to `/usr/local`:
     ```bash
     sudo tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
     ```

3. **Set Up Environment Variables**:
   `bash
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
     `

4. **Apply Changes**:

   - Reload the profile to apply the changes:
     ```bash
     source ~/.bashrc
     ```

5. **Verify Installation**:

   - Check the Go version to confirm installation:
     ```bash
     go version
     ```
   - Expected output:
     ```
     go version go1.24.3 linux/amd64
     ```
 

### Notes:

- Install Docker  
Now, install Docker using the apt package manager:
 ```bash
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io
  ```