# How to Upload a File to Linear

## Include an Image or Video in Markdown Content

The easiest method is to include a URL reference within markdown content when creating issues, comments, or documents:

```graphql
mutation IssueCreate {
    issueCreate(
        input: {
            title: "Issue title"
            description: "Markdown image here: \n ![alt text](https://example.com/image.png)"
            teamId: "9cfb482a-81e3-4154-b5b9-2c805e70a02d"
        }
    ) {
        success
    }
}
```

You can also embed base64 encoded images directly in the markdown.

## Upload Files Manually

To upload files directly to Linear's storage:

### Step 1: Request Upload URL
Use the `fileUpload` mutation to request a pre-signed upload URL:

```graphql
mutation FileUpload {
  fileUpload(
    contentType: "image/png"
    filename: "screenshot.png"
    size: 12345
  ) {
    success
    uploadFile {
      uploadUrl
      assetUrl
      headers {
        key
        value
      }
    }
  }
}
```

### Step 2: Upload to Pre-signed URL
Send a `PUT` request to the returned `uploadUrl` with the file content.

**Note:** Client-side uploads are blocked by Content Security Policy - uploads must be done server-side.

## Example Server-side Upload (TypeScript SDK)

```typescript
async function uploadFileToLinear(file: File): Promise<string> {
  const uploadPayload = await linearClient.fileUpload(file.type, file.name, file.size);
 
  if (!uploadPayload.success || !uploadPayload.uploadFile) {
    throw new Error("Failed to request upload URL");
  }
 
  const uploadUrl = uploadPayload.uploadFile.uploadUrl;
  const assetUrl = uploadPayload.uploadFile.assetUrl;
 
  const headers = new Headers();
  headers.set("Content-Type", file.type);
  headers.set("Cache-Control", "public, max-age=31536000");
  uploadPayload.uploadFile.headers.forEach(({ key, value }) => headers.set(key, value));
 
  try {
    await fetch(uploadUrl, {
      method: "PUT", 
      headers,
      body: file
    });
 
    return assetUrl;
  } catch (e) {
    console.error(e);
    throw new Error("Failed to upload file to Linear");
  }
}
```

## Using Uploaded Files

Once uploaded, you can reference the file using the returned `assetUrl`:

```typescript
// Create issue with uploaded image
const assetUrl = await uploadFileToLinear(file);

await linearClient.createIssue({
  teamId: "team-id",
  title: "Issue with screenshot",
  description: `Here's the screenshot:\n\n![Screenshot](${assetUrl})`
});
```

## Common Errors

### CORS Error
- Client-side uploads are blocked by Content Security Policy
- Solution: Perform uploads server-side

### Authentication Error
- Ensure you're using a valid API key or access token
- Check that the token has appropriate permissions

### File Size Limits
- Check Linear's file size limits for uploads
- Consider compressing large files before upload

## Supported File Types

Linear supports various file types including:
- Images (PNG, JPG, GIF, WebP)
- Videos (MP4, MOV, WebM)
- Documents (PDF)
- Archives (ZIP)

Always specify the correct `contentType` when requesting the upload URL.