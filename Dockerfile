FROM scratch
ADD addressWatcher /opt/addressWatcher/addressWatcher
ENTRYPOINT ["/opt/addressWatcher/addressWatcher"]