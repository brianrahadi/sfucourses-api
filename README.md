<a id="readme-top"></a>

<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![project_license][license-shield]][license-url]

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <!-- <a href="https://github.com/brianrahadi/sfucourses-api">
    <img src="images/logo.png" alt="Logo" width="80" height="80">
  </a> -->

<h3 align="center">sfucourses-api</h3>

  <p align="center">
    REST API server for SFU course outlines, sessions, and instructors
    <br />
    <a href="https://github.com/brianrahadi/sfucourses-api"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/brianrahadi/sfucourses-api">View Demo</a>
    &middot;
    <a href="https://github.com/brianrahadi/sfucourses-api/issues/new?labels=bug&template=bug-report---.md">Report Bug</a>
    &middot;
    <a href="https://github.com/brianrahadi/sfucourses-api/issues/new?labels=enhancement&template=feature-request---.md">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li><a href="#features">Features</a></li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#contact">Contact</a></li>
    <li><a href="#acknowledgments">Acknowledgments</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

[![Product Name Screen Shot][product-screenshot]](https://example.com)
Unofficial API for accessing SFU course outlines, sections, and instructors robustly and used to power sfucourses.com. Data is pulled from SFU Course Outlines REST API. This API is not affiliated with Simon Fraser University.

See [api.sfucourses.com](https://api.sfucourses.com)
<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

- Golang

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->

## Getting Started

To get a local copy up and running follow these simple example steps.

### Prerequisites

- Golang v1.23.3

### Quick Features

- REST API Server - [api.sfucourses.com](https://api.sfucourses.com)
- Golang Script to fetch outlines, sessions, and sync instructors

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/brianrahadi/sfucourses-api.git
   ```
2. Change git remote url to avoid accidental pushes to base project
   ```sh
   git remote set-url origin brianrahadi/sfucourses-api
   git remote -v
   ```

3. Run the project through air.toml or docker

air is good for development with it's real-time file update sync
```
air
```

Docker is good for its 1-to-1 behaviour with production. You can either use docker build and run or docker compose.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Docker

### Build the Image
```bash
docker build -t sfu-courses-api .
```

### Run the Container
```bash
docker run -p 8080:8080 sfu-courses-api
```

### Useful Docker Commands
```bash
# View running containers
docker ps

# View logs
docker logs <container_id>

# View logs live
docker logs -f <container_id>

# Stop container
docker stop <container_id>

# Remove container
docker rm <container_id>
```

The API will be available at `http://localhost:8080` once the container is running.

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/brianrahadi/sfucourses-api/issues) for a full list of proposed features (and known issues).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->

## Contributing
Very recommended! very appreciated!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Docker Commands

```


### UPDATE_PASSWORD
Required for the `/update` endpoint. Set this environment variable to secure manual data updates.

```bash
export UPDATE_PASSWORD="your-secure-password-here"
```

**Usage:**
```bash
curl -X POST http://localhost:8080/update \
  -H "Content-Type: text/plain" \
  -d "your-secure-password-here"
```

**Security Note:** Never commit the actual password to version control. Use environment variables or secrets management in production.

### Git Hooks
To set up the pre-commit hooks:
```bash
cp hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```


This way:
- The hook templates are version controlled
- Each developer can set up their own hooks
- The actual `.git/hooks` directory remains local to each developer's machine

<!-- LICENSE -->

## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->

## Contact

Brian Rahadi - brian.rahadi@gmail.com

Project Link: [https://github.com/brianrahadi/sfucourses-api](https://github.com/brianrahadi/sfucourses-api)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- ACKNOWLEDGMENTS -->

## Acknowledgments

- https://go.dev/doc/ - Golang Dev docs

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[contributors-shield]: https://img.shields.io/github/contributors/brianrahadi/sfucourses-api.svg?style=for-the-badge
[contributors-url]: https://github.com/brianrahadi/sfucourses-api/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/brianrahadi/sfucourses-api.svg?style=for-the-badge
[forks-url]: https://github.com/brianrahadi/sfucourses-api/network/members
[stars-shield]: https://img.shields.io/github/stars/brianrahadi/sfucourses-api.svg?style=for-the-badge
[stars-url]: https://github.com/brianrahadi/sfucourses-api/stargazers
[issues-shield]: https://img.shields.io/github/issues/brianrahadi/sfucourses-api.svg?style=for-the-badge
[issues-url]: https://github.com/brianrahadi/sfucourses-api/issues
[license-shield]: https://img.shields.io/github/license/brianrahadi/sfucourses-api.svg?style=for-the-badge
[license-url]: https://github.com/brianrahadi/sfucourses-api/blob/master/LICENSE.txt
[product-screenshot]: images/screenshot.png
[Next.js]: https://img.shields.io/badge/next.js-000000?style=for-the-badge&logo=nextdotjs&logoColor=white
[Next-url]: https://nextjs.org/
[React.js]: https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB
[React-url]: https://reactjs.org/
[Vue.js]: https://img.shields.io/badge/Vue.js-35495E?style=for-the-badge&logo=vuedotjs&logoColor=4FC08D
[Vue-url]: https://vuejs.org/
[Angular.io]: https://img.shields.io/badge/Angular-DD0031?style=for-the-badge&logo=angular&logoColor=white
[Angular-url]: https://angular.io/
[Svelte.dev]: https://img.shields.io/badge/Svelte-4A4A55?style=for-the-badge&logo=svelte&logoColor=FF3E00
[Svelte-url]: https://svelte.dev/
[Laravel.com]: https://img.shields.io/badge/Laravel-FF2D20?style=for-the-badge&logo=laravel&logoColor=white
[Laravel-url]: https://laravel.com
[Bootstrap.com]: https://img.shields.io/badge/Bootstrap-563D7C?style=for-the-badge&logo=bootstrap&logoColor=white
[Bootstrap-url]: https://getbootstrap.com
[JQuery.com]: https://img.shields.io/badge/jQuery-0769AD?style=for-the-badge&logo=jquery&logoColor=white
[JQuery-url]: https://jquery.com
