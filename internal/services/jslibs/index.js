// /home/akash/API2SDK/internal/services/jslibs/index.js
// @ts-check
import p2o from 'postman-to-openapi';

// Expose the function directly on globalThis with a unique name
// @ts-ignore
globalThis.myCustomP2OFunction = async (collectionString, optionsString) => {
  try {
    console.log("[P2O JS] Received collection string (first 100 chars):", collectionString.substring(0, 100));
    console.log("[P2O JS] Received options string:", optionsString);

    const collectionObject = JSON.parse(collectionString);
    console.log("[P2O JS] Parsed collection string to object.");

    let options = optionsString ? JSON.parse(optionsString) : {}; // Changed const to let
    // Ensure replaceVars is true to handle {{baseUrl}} and other Postman variables
    options.replaceVars = true;
    // Attempt to use Postman request names for operationIds
    options.operationId = 'OPERATION_NAME';

    // Attempt to explicitly set server information
    if (collectionObject && collectionObject.variable) {
      const baseUrlVar = collectionObject.variable.find(v => v.key === 'baseUrl');
      if (baseUrlVar && baseUrlVar.value) {
        options.servers = [{ url: baseUrlVar.value }];
        console.log("[P2O JS] Added explicit server URL to options:", baseUrlVar.value);
      }
    }
    
    console.log("[P2O JS] Parsed and modified options:", JSON.stringify(options)); // Modified log message

    console.log("[P2O JS] Calling p2o with RAW COLLECTION STRING and NULL for output path."); // Modified log
    const result = await p2o(collectionString, null, options); // Pass collectionString directly
    
    console.log("[P2O JS] Conversion successful. Result type:", typeof result);
    if (typeof result === 'string') {
        console.log("[P2O JS] Conversion result (first 200 chars):", result.substring(0, 200));
    } else {
        console.log("[P2O JS] Conversion result (not a string):", result);
    }
    return result;
  } catch (e) {
    console.error("[P2O JS] Error during conversion:", e);
    if (e.stack) {
        console.error("[P2O JS] Error stack:", e.stack);
    }
    // Ensure a serializable error is thrown back to Go
    throw new Error(`JavaScript conversion failed: ${e.message || String(e)}`);
  }
};

console.log("js: myCustomP2OFunction has been set on globalThis.");
// @ts-ignore
if (typeof globalThis.myCustomP2OFunction === 'function') {
  console.log("js: typeof globalThis.myCustomP2OFunction is function.");
} else {
  console.error("js: typeof globalThis.myCustomP2OFunction is NOT a function.");
}
