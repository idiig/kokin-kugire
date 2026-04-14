;;; .dir-locals.el — local Emacs settings for docs/

((org-mode
  . ((eval
      . (defun kugire-short-manuscript ()
          "Write a short version of the manuscript to manuscript-short.org.
Keeps the #+title, all first-level headings, and the TL;DR
#+begin_comment block that opens each section.  The original
buffer is not modified.  Output file is written next to the
source file and opened in a new window."
          (interactive)
          (let* ((src-buf (current-buffer))
                 (out-file (expand-file-name
                            "manuscript-short.org"
                            (file-name-directory (buffer-file-name src-buf))))
                 title
                 sections)
            ;; ── collect title ────────────────────────────────────────
            (with-current-buffer src-buf
              (save-excursion
                (goto-char (point-min))
                (when (re-search-forward "^#\\+title:[ \t]*\\(.*\\)$" nil t)
                  (setq title (match-string-no-properties 1))))
              ;; ── collect headings + TL;DR comment-blocks ─────────────
              ;; #+begin_comment...#+end_comment is parsed as 'comment-block
              ;; (not 'special-block); its text is in :value.
              (org-element-map (org-element-parse-buffer) 'headline
                (lambda (hl)
                  (when (= (org-element-property :level hl) 1)
                    (let ((heading (org-element-property :raw-value hl))
                          (tldr nil))
                      (org-element-map (org-element-contents hl) 'comment-block
                        (lambda (blk)
                          (setq tldr
                                (replace-regexp-in-string
                                 "^TL;DR:[ \t]*" ""
                                 (string-trim
                                  (org-element-property :value blk)))))
                        nil t) ; first match only
                      (push (cons heading tldr) sections))))))
            ;; ── write output file ────────────────────────────────────
            (with-temp-file out-file
              (org-mode)
              (when title
                (insert (format "#+title: %s\n\n" title)))
              (dolist (sec (nreverse sections))
                (insert (format "* %s\n\n" (car sec)))
                (when (cdr sec)
                  (insert (cdr sec))
                  (insert "\n"))
                (insert "\n")))
            (find-file-other-window out-file)
            (goto-char (point-min))
            (message "Short manuscript written to %s" out-file)))))))
