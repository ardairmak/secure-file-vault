# 🔐 Secure File Vault

## Overview

Secure File Vault is an advanced, cross-platform file encryption application designed to protect your sensitive documents with enhanced security and an intuitive user interface. This project is developed as part of a Computer Security lecture to demonstrate secure file handling and encryption techniques.

## 🌟 Key Features

- **🔒 Advanced Encryption**: AES-256 encryption protecting your files
- **🔑 Robust Authentication**: Secure password management with Scrypt
- **🛡️ Integrity Verification**: SHA-256 hash checks to detect file tampering
- **🖥️ Cross-Platform**: Works seamlessly on Windows, macOS, and Linux
- **📂 Intuitive File Management**: User-friendly Fyne GUI
- **🕵️ Real-Time File Monitoring**: Automatic change detection and vault synchronization

## 📦 Prerequisites

- Go 1.22+
- Git
- Fyne dependencies

## 🚀 Quick Start

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/ardairmak/secure-file-vault.git
   cd secure-file-vault
   ```

2. **Install Go Dependencies**

   ```bash
   go mod download
   ```

3. **Install Fyne CLI**

   ```bash
   go install fyne.io/fyne/v2/cmd/fyne@latest
   ```

4. **Bundle Assets**

   Navigate to the `ui` directory and run the asset bundling script:

   ```bash
   cd ui
   chmod +x bundle_assets.sh
   ./bundle_assets.sh
   cd ..
   ```

5. **Build the Application**

   ```bash
   go build -o secure-file-vault ./main.go
   ```

6. **Run the Application**

   ```bash
   ./secure-file-vault
   ```

## Usage

### Creating an Account and Vault

- **Launch the Application**: Upon starting, you'll see the login screen.
- **Register**: Click on "Not a member? Register here." to create a new account.
- **Fill in Details**: Enter a username and a strong password.
- **Vault Path**: Specify a custom vault path or leave it blank to use the default location.
- **Register**: Click the "Register" button to create your account and vault.

### Adding Files to the Vault

- **Access Main Screen**: After logging in, you'll be on the main screen.
- **Select File**: Click "Select File" to choose a file from your system.
- **Add File**: After selecting, click "Add File" to encrypt and add it to your vault.

### Viewing and Managing Files

- **View Files**: Click on "View Files" to see a list of files stored in your vault.
- **Select Files**: Use the checkboxes to select files for actions.

### Extracting Files

- **Select Files**: In the file list, select the files you wish to extract.
- **Extract**: Click the "Extract" button and choose a destination folder.
- **Monitoring**: Extracted files are monitored for changes and can be updated back into the vault.

### Updating Files

- **Modify Extracted File**: Make changes to the extracted file as needed.
- **Automatic Prompt**: On detecting changes, the application prompts you to update the vault.
- **Update Vault**: Confirm to encrypt the updated file and save it back into the vault.

### Locking and Unlocking the Vault

- **Lock Vault**: Log out by clicking the "Logout" button to lock the vault.
- **Unlock Vault**: Log in with your credentials to unlock and access your files.

## 🔐 Security

**Disclaimer**: While this application provides robust security, it is not recommended for protecting secret classified information without additional security audits.

### Security Features

- AES-256 encryption
- Scrypt key derivation
- SHA-256 integrity checks
- Secure file deletion
- No plaintext password storage

## 🛠️ Development

### Packaging with Fyne

Fyne provides tools to package your application for different operating systems.

**For macOS:**

```bash
fyne package -os darwin -icon logo.png
```

- This command creates a `.app` bundle that macOS users can run directly.

**For Windows:**

```bash
fyne package -os windows -icon logo.png
```

- This generates an `.exe` file along with necessary resources.

**For Linux:**

```bash
fyne package -os linux -icon logo.png
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
