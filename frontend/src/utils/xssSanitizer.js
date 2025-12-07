// frontend/src/utils/xssSanitizer.js

/**
 * XSS Sanitization Library
 * A comprehensive utility for protecting against cross-site scripting (XSS) attacks.
 * This module provides various functions to sanitize user inputs and prevent malicious code injection.
 */

/**
 * Escapes HTML special characters to prevent them from being interpreted as markup.
 * Transforms dangerous characters like '<' into their harmless HTML entity equivalents.
 * This is the foundation of our XSS protection strategy - neutralizing potentially executable code.
 *
 * @param {string} unsafe - The raw, potentially dangerous string
 * @returns {string} - The escaped, safe string where HTML special chars are converted to entities
 */
export function escapeHtml(unsafe) {
  return unsafe
    .replace(/&/g, "&amp;")     // & → &amp;  (ampersand)
    .replace(/</g, "&lt;")      // < → &lt;   (less than)
    .replace(/>/g, "&gt;")      // > → &gt;   (greater than)
    .replace(/"/g, "&quot;")    // " → &quot; (double quote)
    .replace(/'/g, "&#039;");   // ' → &#039; (single quote)
}

/**
 * Sanitizes plain text before outputting it in an HTML context.
 * Acts as our front line of defense against XSS in text contexts.
 * If input isn't a string, returns it unchanged (handles numbers, booleans, etc. safely).
 *
 * @param {any} text - Text to sanitize, typically user-provided input
 * @returns {any} - Sanitized text or the original input if not a string
 */
export function sanitizeText(text) {
  if (typeof text !== 'string') {
    return text;
  }
  return escapeHtml(text);
}

/**
 * Sanitizes URLs to prevent javascript: protocol exploits.
 * URLs are a common attack vector where malicious javascript: URLs can execute code.
 * This function ensures URLs use safe protocols by filtering out dangerous ones.
 *
 * @param {string} url - The URL to sanitize
 * @returns {string} - Sanitized URL or empty string if dangerous
 */
export function sanitizeUrl(url) {
  if (!url) return '';

  // Remove whitespace that might be used to obfuscate malicious URLs
  const trimmedUrl = url.trim();

  // Check for dangerous protocols that could execute code
  const dangerous = /^(javascript|data|vbscript|file):/i;
  if (dangerous.test(trimmedUrl)) {
    return '';
  }

  return trimmedUrl;
}

/**
 * Recursively sanitizes an entire data object (or array).
 * The nuclear option - sanitizes everything in a complex data structure.
 * Walks through objects and arrays, sanitizing all string values it finds.
 *
 * @param {any} data - The data structure to sanitize
 * @returns {any} - A new sanitized copy of the input data structure
 */
export function sanitizeData(data) {
  if (!data) return data;

  // Sanitize strings
  if (typeof data === 'string') {
    return sanitizeText(data);
  }

  // Recursively sanitize arrays
  if (Array.isArray(data)) {
    return data.map(item => sanitizeData(item));
  }

  // Recursively sanitize objects
  if (typeof data === 'object' && data !== null) {
    const sanitized = {};
    for (const [key, value] of Object.entries(data)) {
      sanitized[key] = sanitizeData(value);
    }
    return sanitized;
  }

  // Other types (numbers, booleans, etc.) pass through unchanged
  return data;
}

/**
 * Vue directive for safely rendering HTML content.
 * Allows controlled injection of HTML while filtering out dangerous elements and attributes.
 * Use this when you absolutely need to render user-provided HTML (but remember, this should be rare!).
 *
 * Usage example: <div v-safe-html="htmlContent"></div>
 */
export const safeHtml = {
  mounted(el, binding) {
    // Define allowed HTML elements and attributes (whitelist approach)
    const allowedTags = ['p', 'br', 'b', 'i', 'strong', 'em', 'ul', 'ol', 'li', 'span'];
    const allowedAttributes = ['class', 'id', 'style'];

    let html = binding.value || '';

    // Parse the HTML string into a DOM tree for safer manipulation
    const parser = new DOMParser();
    const doc = parser.parseFromString(html, 'text/html');

    /**
     * Recursively sanitizes a DOM node and its children.
     * The heavy lifter - examines each node and attribute against our whitelist.
     *
     * @param {Node} node - DOM node to sanitize
     * @returns {Node} - Clean, sanitized node
     */
    function sanitizeNode(node) {
      // Text nodes are always safe - pass through unchanged
      if (node.nodeType === Node.TEXT_NODE) {
        return node.cloneNode(true);
      }

      // Process element nodes (actual HTML tags)
      if (node.nodeType === Node.ELEMENT_NODE) {
        // If tag isn't in our whitelist, replace with a safe 'span'
        const newNode = allowedTags.includes(node.tagName.toLowerCase())
          ? document.createElement(node.tagName)
          : document.createElement('span');

        // Copy only allowed attributes
        Array.from(node.attributes).forEach(attr => {
          if (allowedAttributes.includes(attr.name)) {
            newNode.setAttribute(attr.name, attr.value);
          }
        });

        // Recursively process all child nodes
        Array.from(node.childNodes).forEach(child => {
          const safeChild = sanitizeNode(child);
          newNode.appendChild(safeChild);
        });

        return newNode;
      }

      // For any other node types, return an empty text node (completely harmless)
      return document.createTextNode('');
    }

    // Sanitize the entire body element and its contents
    const safeBody = sanitizeNode(doc.body);

    // Replace the element's content with sanitized version
    el.innerHTML = '';
    Array.from(safeBody.childNodes).forEach(node => {
      el.appendChild(node);
    });
  }
};
