/* Fix for loop in 0120_2.9.0_schema.up.sql that never populated fixable_cnt.*/
DO
$$
    DECLARE
        report        RECORD;
        v             RECORD;
        fixable_count BIGINT;
    BEGIN
        FOR report IN SELECT uuid FROM scan_report WHERE fixable_cnt IS NULL
            LOOP
                fixable_count := 0;
                FOR v IN SELECT vr.fixed_version
                         FROM report_vulnerability_record rvr,
                              vulnerability_record vr
                         WHERE rvr.report_uuid = report.uuid
                           AND rvr.vuln_record_id = vr.id
                    LOOP
                        IF v.fixed_version IS NOT NULL AND v.fixed_version != '' THEN
                            fixable_count := fixable_count + 1;
                        END IF;
                    END LOOP;
                UPDATE scan_report
                SET fixable_cnt = fixable_count
                WHERE uuid = report.uuid;
            END LOOP;
    END
$$;
