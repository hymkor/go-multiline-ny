; https://github.com/hymkor/lispect

(defconstant example "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890")

(defglobal pid (spawn "go" "run" "slow.go"))
(expect "[1]")
(defglobal start (get-internal-real-time))
(send example)
(send #\UA)
(send #\U4)
(wait pid)
(defglobal end (get-internal-real-time))
(format (standard-output) "INPUT TIME: ~a~%" (div (convert (- end start) <float>) (internal-time-units-per-second)))

(defglobal pid (spawn "go" "run" "slow.go"))
(expect "[1]")
(setq start (get-internal-real-time))
(send example)
(send #\UA)
(send #\U4)
(sendln (create-string (length example) #\U2))
(wait pid)
(setq end (get-internal-real-time))
(format (standard-output) "MOVING TIME: ~a~%" (div (convert (- end start) <float>) (internal-time-units-per-second)))
