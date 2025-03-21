# Golang GraphQL API with Fiber & BigCache

1. This project has **GraphQL service** that fetches materials from an external **REST API**.  
   Since the material API returns different results on each request, we call it **5 times** to collect all available materials.
2. The **GraphQL logic** is fully implemented inside the `api/graphql` folder, making it modular and **future-ready for gRPC or REST API integration**.

## Tech Stack

- **Golang**: Backend language
- **Fiber**: Fast HTTP framework
- **GraphQL**: Query language for APIs

---

## Features

**GraphQL API** with Fiber  
 **BigCache integration** to cache responses

---

## Installation & Setup

### **1️ Clone the Repository**

git clone https://github.com/medigini/task.git
cd your-repo
