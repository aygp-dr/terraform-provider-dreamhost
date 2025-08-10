terraform {
  required_providers {
    dreamhost = {
      version = "0.0.1"
      source  = "hashicorp.com/edu/dreamhost"
    }
  }
}

provider "dreamhost" {
  api_key = ""
}
