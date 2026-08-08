# Project Customization & Server Notes

## Server Connection Info
- **Host**: `look-victory-listings-roland.trycloudflare.com` (dynamic trycloudflare host)
- **User**: `osboxes`
- **Password**: `liverfag@P3`
- **SSH Command**:
  ```bash
  ssh -o "ProxyCommand=cloudflared access tcp --hostname %h" osboxes@look-victory-listings-roland.trycloudflare.com
  ```
  *(Tip: Add `-o StrictHostKeyChecking=accept-new` for dynamic trycloudflare URLs)*
