// frontend/src/services/imageUtils.js
/**
 * A collection of magical tools for image manipulation and optimization.
 * Because pushing raw 8K cat pictures to your server is just being rude.
 */

/**
 * The Grand Image Optimizer - shrinks, compresses, transforms images, and crops to square!
 * Takes a user-uploaded behemoth and turns it into a civilized web-friendly asset.
 * Like a tailor who takes your oversized suit and makes it fit just right,
 * this function makes sure images aren't consuming bandwidth like it's free.
 * Now with special square-cropping powers for perfect avatars!
 *
 * @param {File|Blob} imageFile - The untamed image that needs domestication
 * @param {Object} options - Your tailoring instructions
 * @param {Number} options.maxWidth - Maximum allowed width before intervention (default: 800px)
 * @param {Number} options.maxHeight - Maximum allowed height before intervention (default: 800px)
 * @param {Number} options.quality - JPEG quality setting (0-1, default: 0.8)
 * @param {String} options.format - Desired image format ('jpeg', 'png', 'webp')
 * @param {Boolean} options.cropToSquare - Whether to crop the image to a square (default: false)
 * @returns {Promise<{blob: Blob, base64: String}>} - The now well-behaved image
 */
export async function optimizeImage(imageFile, options = {}) {
  const {
    maxWidth = 800,
    maxHeight = 800,
    quality = 0.8,
    format = 'jpeg',
    cropToSquare = false // New option to force square crop
  } = options;

  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = (e) => {
      const img = new Image();

      img.onload = () => {
        // Original dimensions
        const origWidth = img.width;
        const origHeight = img.height;
        
        // Canvas for our manipulations
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        
        // Handle square cropping if requested - perfect for avatars!
        if (cropToSquare) {
          // Determine the size of our square (smallest dimension)
          const squareSize = Math.min(origWidth, origHeight);
          
          // Calculate crop coordinates to take center portion
          const cropX = (origWidth - squareSize) / 2;
          const cropY = (origHeight - squareSize) / 2;
          
          // Set canvas to be a perfect square
          canvas.width = squareSize;
          canvas.height = squareSize;
          
          // Draw only the center portion to the canvas
          ctx.drawImage(
            img,
            cropX, cropY, squareSize, squareSize, // Source rectangle
            0, 0, squareSize, squareSize           // Destination rectangle
          );
          
          // Now resize the square if it's too large
          let finalSize = squareSize;
          if (finalSize > maxWidth || finalSize > maxHeight) {
            finalSize = Math.min(maxWidth, maxHeight);
            
            // Create a temporary canvas for resizing
            const tempCanvas = document.createElement('canvas');
            tempCanvas.width = finalSize;
            tempCanvas.height = finalSize;
            const tempCtx = tempCanvas.getContext('2d');
            
            // Draw the square image at a smaller size
            tempCtx.drawImage(canvas, 0, 0, finalSize, finalSize);
            
            // Replace our working canvas
            canvas.width = finalSize;
            canvas.height = finalSize;
            ctx.drawImage(tempCanvas, 0, 0);
          }
        } else {
          // If not cropping to square, follow the original resizing logic
          let width = origWidth;
          let height = origHeight;

          if (width > height) {
            if (width > maxWidth) {
              height = Math.round(height * maxWidth / width);
              width = maxWidth;
            }
          } else {
            if (height > maxHeight) {
              width = Math.round(width * maxHeight / height);
              height = maxHeight;
            }
          }

          canvas.width = width;
          canvas.height = height;
          ctx.drawImage(img, 0, 0, width, height);
        }

        // Determine MIME type
        let mimeType;
        switch (format.toLowerCase()) {
          case 'png':
            mimeType = 'image/png';
            break;
          case 'webp':
            mimeType = 'image/webp';
            break;
          case 'jpeg':
          case 'jpg':
          default:
            mimeType = 'image/jpeg';
            break;
        }

        // Convert to base64 and blob
        const base64 = canvas.toDataURL(mimeType, quality);

        // Convert base64 to blob for server transmission
        canvas.toBlob((blob) => {
          resolve({
            blob,
            base64,
            width: canvas.width,
            height: canvas.height,
            format: mimeType
          });
        }, mimeType, quality);
      };

      img.onerror = () => {
        reject(new Error('Nepavyko įkelti nuotraukos'));
      };

      img.src = e.target.result;
    };

    reader.onerror = () => {
      reject(new Error('Nepavyko nuskaityti failo'));
    };

    reader.readAsDataURL(imageFile);
  });
}

/**
 * The Miniature Artist - creates tiny versions of images with elegance!
 * Generates thumbnails for previews, avatars, and other small image needs.
 * Like a portrait painter specializing in miniatures, this function
 * creates small but perfectly formed versions of the original image.
 *
 * @param {File|Blob} imageFile - The full-size masterpiece to be miniaturized
 * @param {Object} options - Artistic directions for the miniature
 * @returns {Promise<String>} - Base64 representation of the thumbnail
 */
export async function createThumbnail(imageFile, options = {}) {
  const {
    width = 150,
    height = 150,
    quality = 0.7,
    format = 'jpeg',
    cropToSquare = true // Default to square cropping for thumbnails/avatars
  } = options;

  const result = await optimizeImage(imageFile, {
    maxWidth: width,
    maxHeight: height,
    quality,
    format,
    cropToSquare
  });

  return result.base64;
}

/**
 * The Image Bouncer - checks if files meet the entry requirements!
 * Validates image files against size and format restrictions.
 * Like a nightclub doorman with a very specific guest list,
 * this function decides which images are allowed into your application.
 *
 * @param {File} file - The file hoping to gain entry to your application
 * @param {Object} constraints - The VIP list requirements
 * @returns {Object} - Validation result with status and explanation
 */
export function validateImage(file, constraints = {}) {
  const {
    maxSizeKB = 5120, // 5MB
    allowedTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/gif']
  } = constraints;

  if (!file) {
    return { valid: false, message: 'Failas nepasirinktas' };
  }

  if (!allowedTypes.includes(file.type)) {
    return { valid: false, message: 'Nepalaikomas failo formatas' };
  }

  if (file.size > maxSizeKB * 1024) {
    return { valid: false, message: `Failo dydis viršija ${maxSizeKB / 1024} MB` };
  }

  return { valid: true, message: 'Failas tinkamas' };
}

/**
 * The Image Precognition Specialist - sees the future by loading images in advance!
 * Preloads images in the background to prevent visual hiccups when they're needed.
 * Like preparing your umbrella before the rain starts, this function
 * ensures images are ready to display the moment they're required.
 *
 * @param {String} src - The URL location of the image to be summoned from the future
 * @returns {Promise<HTMLImageElement>} - Promise with the preloaded image element
 */
export function preloadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error(`Nepavyko įkelti nuotraukos: ${src}`));
    img.src = src;
  });
}
