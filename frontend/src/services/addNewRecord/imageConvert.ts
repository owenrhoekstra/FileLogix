export const convertToWebP = (file: File, quality = 0.92): Promise<File> => {
    return new Promise((resolve, reject) => {
        const img = new Image()
        const url = URL.createObjectURL(file)

        img.onload = () => {
            const canvas = document.createElement('canvas')
            canvas.width = img.naturalWidth
            canvas.height = img.naturalHeight

            const ctx = canvas.getContext('2d')
            if (!ctx) {
                URL.revokeObjectURL(url)
                return reject(new Error('Failed to get canvas context'))
            }

            ctx.drawImage(img, 0, 0)
            URL.revokeObjectURL(url)

            canvas.toBlob(
                (blob) => {
                    if (!blob) return reject(new Error('Failed to convert image to WebP'))
                    const converted = new File(
                        [blob],
                        file.name.replace(/\.[^.]+$/, '.webp'),
                        { type: 'image/webp' }
                    )
                    resolve(converted)
                },
                'image/webp',
                quality
            )
        }

        img.onerror = () => {
            URL.revokeObjectURL(url)
            reject(new Error(`Failed to load image: ${file.name}`))
        }

        img.src = url
    })
}

export const convertAllToWebP = async (files: FileList | File[], quality = 0.92): Promise<File[]> => {
    return Promise.all(Array.from(files).map(f => convertToWebP(f, quality)))
}