# Use an official Node.js runtime as the base image
FROM node:20

# Set the working directory in the Docker container
WORKDIR /app

# Copy package.json and package-lock.json to the working directory
COPY package*.json ./

# Copy the entire project to the working directory
COPY . .

# Expose the port the app will run on
EXPOSE 3000

# Run the app when the Docker image is run
CMD ["node", "index.js"]

