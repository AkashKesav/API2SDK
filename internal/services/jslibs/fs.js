// /home/akash/API2SDK/internal/services/jslibs/fs.js
// Minimal fs shim for postman-to-openapi, assuming it only needs promises.readFile/writeFile
// and that these won't actually be called if input/output are strings.
console.log('[API2SDK FS Shim] fs module loaded');
module.exports = {
  promises: {
    writeFile: async (filePath, data) => {
      console.warn(`[API2SDK FS Shim] fs.promises.writeFile called for ${filePath}. This should ideally not happen.`);
      // In a non-Node environment, this can't do anything.
      // If this function is critical, the conversion will fail or produce incorrect results.
    },
    readFile: async (filePath) => {
      const pathStr = String(filePath); // Attempt to convert path to string

      // If the "path" is actually our direct JSON input (which is an object, not a path string),
      // or if it's an empty string (which we pass when we intend to provide content directly),
      // it means the library is trying to read the content it already has.
      // This is a common pattern if a library can accept a path OR direct content.
      // However, postman-to-openapi seems to always try readFile if a string is passed,
      // even if that string is the content itself.
      if (typeof filePath !== 'string') { // If it's not a string (e.g., an object)
        try {
          const jsonContent = JSON.stringify(filePath);
          console.warn(`[API2SDK FS Shim] fs.promises.readFile called with an object. Serialized and returning as JSON string. Content (first 100): ${jsonContent.substring(0,100)}`);
          return Promise.resolve(jsonContent);
        } catch (e) {
          console.error(`[API2SDK FS Shim] fs.promises.readFile called with an object that could not be serialized: ${pathStr}`, e);
          return Promise.reject(new Error(`[API2SDK FS Shim] Could not serialize object passed to readFile: ${e.message}`));
        }
      }
      
      if (filePath === '' || pathStr.startsWith('{') || pathStr.startsWith('[')) { // If it's a string but looks like content
        console.warn(`[API2SDK FS Shim] fs.promises.readFile called with string content that looks like JSON: '${pathStr.substring(0,100)}...'. Returning content directly.`);
        // This case might be hit if p2o is given a string and tries to read it.
        return Promise.resolve(pathStr); 
      }

      console.error(`[API2SDK FS Shim] fs.promises.readFile: True file system access attempted for path '${pathStr}' and is not available in this environment.`);
      return Promise.reject(new Error(`[API2SDK FS Shim] File system access not available for path: ${pathStr}`));
    },
    // Add other fs functions if they are reported as missing by the library
    // For example, existsSync, statSync, etc.
    // stat: async (path) => {
    //   console.warn(`[API2SDK FS Shim] fs.promises.stat called for ${String(path)}. Returning dummy stats.`);
    //   return Promise.resolve({
    //     isFile: () => true,
    //     isDirectory: () => false,
    //     isBlockDevice: () => false,
    //     isCharacterDevice: () => false,
    //     isSymbolicLink: () => false,
    //     isFIFO: () => false,
    //     isSocket: () => false,
    //     size: 0,
    //   });
    // },
    // access: async (path, mode) => {
    //   console.warn(`[API2SDK FS Shim] fs.promises.access called for ${String(path)}. Assuming accessible.`);
    //   return Promise.resolve();
    // }
  },
  // Synchronous stubs if needed, though esbuild browser platform should handle some of this.
  // readFileSync: (path, options) => {
  //   console.error(`[API2SDK FS Shim] fs.readFileSync: File system access not available for path: ${String(path)}`);
  //   throw new Error(`[API2SDK FS Shim] File system access not available for path: ${String(path)} (readFileSync)`);
  // },
  // writeFileSync: (path, data, options) => {
  //   console.warn(`[API2SDK FS Shim] fs.writeFileSync called for path '${String(path)}'. This operation is not supported and will be a no-op.`);
  // },
};

console.log("[API2SDK FS Shim] fs module (readFile/writeFile promises) shim loaded with improved logic.");
