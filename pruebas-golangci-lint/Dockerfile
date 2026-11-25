FROM golang:1.27.1-trixie
WORKDIR /lab
COPY repo/ /lab/repo/
COPY repo-arreglado/ /lab/repo-arreglado/
COPY repo-mas-bugs/ /lab/repo-mas-bugs/
COPY repo-falla-real/ /lab/repo-falla-real/
COPY dentro.sh /lab/dentro.sh
RUN chmod +x /lab/dentro.sh
CMD ["/lab/dentro.sh"]
