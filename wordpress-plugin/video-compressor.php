<?php
/**
 * Plugin Name: Video Compressor API Integration
 * Plugin URI: https://api.trendss.net
 * Description: Automatically compress videos and images using external compression API
 * Version: 1.0.0
 * Author: Your Name
 * Author URI: https://yourwebsite.com
 * License: GPL v2 or later
 * Text Domain: video-compressor-api
 */

defined('ABSPATH') || exit;

class VideoCompressorAPI {
    
    private $api_url;
    private $api_key;
    
    public function __construct() {
        // Configuration
        $this->api_url = 'https://api.trendss.net/api';
        $this->api_key = 'sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1';
        
        // Hooks
        add_action('add_attachment', [$this, 'auto_compress_media']);
        add_action('admin_menu', [$this, 'add_admin_menu']);
        add_action('admin_enqueue_scripts', [$this, 'enqueue_scripts']);
        
        // AJAX handlers
        add_action('wp_ajax_compress_media', [$this, 'ajax_compress_media']);
        add_action('wp_ajax_check_compression_status', [$this, 'ajax_check_status']);
        add_action('wp_ajax_get_compression_result', [$this, 'ajax_get_result']);
        
        // Settings
        add_action('admin_init', [$this, 'register_settings']);
        
        // Cron job for checking status
        add_action('compressor_check_pending_jobs', [$this, 'check_pending_jobs']);
        if (!wp_next_scheduled('compressor_check_pending_jobs')) {
            wp_schedule_event(time(), 'hourly', 'compressor_check_pending_jobs');
        }
    }
    
    /**
     * Register settings
     */
    public function register_settings() {
        register_setting('video_compressor_options', 'vc_api_url');
        register_setting('video_compressor_options', 'vc_api_key');
        register_setting('video_compressor_options', 'vc_auto_compress');
        register_setting('video_compressor_options', 'vc_default_quality');
    }
    
    /**
     * Auto-compress media when uploaded
     */
    public function auto_compress_media($attachment_id) {
        $auto_compress = get_option('vc_auto_compress', true);
        
        if (!$auto_compress) {
            return;
        }
        
        $mime_type = get_post_mime_type($attachment_id);
        
        // Determine compression type
        $compression_type = null;
        if (strpos($mime_type, 'video/') === 0) {
            $compression_type = 'video';
        } elseif (strpos($mime_type, 'image/') === 0) {
            $compression_type = 'image';
        } else {
            return; // Not video or image
        }
        
        // Compress
        $this->compress_media($attachment_id, $compression_type);
    }
    
    /**
     * Compress media
     */
    public function compress_media($attachment_id, $compression_type = 'video') {
        $file_url = wp_get_attachment_url($attachment_id);
        $quality = get_option('vc_default_quality', 'medium');
        
        $data = [
            'post_id' => $attachment_id,
            'compression_type' => $compression_type,
            'priority' => 5
        ];
        
        if ($compression_type === 'video' || $compression_type === 'both') {
            $data['video_data'] = [
                'file_url' => $file_url,
                'quality' => $quality,
                'hls_enabled' => false
            ];
        }
        
        if ($compression_type === 'image' || $compression_type === 'both') {
            $data['image_data'] = [
                'file_url' => $file_url,
                'quality' => $quality,
                'variants' => ['thumbnail', 'medium', 'large']
            ];
        }
        
        $response = wp_remote_post($this->api_url . '/compress', [
            'headers' => [
                'X-API-Key' => $this->api_key,
                'Content-Type' => 'application/json',
            ],
            'body' => json_encode($data),
            'timeout' => 30
        ]);
        
        if (is_wp_error($response)) {
            error_log('[Video Compressor] API Error: ' . $response->get_error_message());
            return false;
        }
        
        $status_code = wp_remote_retrieve_response_code($response);
        $body = json_decode(wp_remote_retrieve_body($response), true);
        
        if ($status_code === 200 && isset($body['job_id'])) {
            // Save job info
            update_post_meta($attachment_id, '_compression_job_id', $body['job_id']);
            update_post_meta($attachment_id, '_compression_status', 'queued');
            update_post_meta($attachment_id, '_compression_type', $compression_type);
            update_post_meta($attachment_id, '_compression_queue_position', $body['queue_position']);
            update_post_meta($attachment_id, '_compression_estimated_time', $body['estimated_time']);
            
            return $body['job_id'];
        }
        
        error_log('[Video Compressor] API Error: ' . print_r($body, true));
        return false;
    }
    
    /**
     * Get job status
     */
    public function get_status($job_id) {
        $response = wp_remote_get($this->api_url . '/status/' . $job_id, [
            'headers' => ['X-API-Key' => $this->api_key],
            'timeout' => 15
        ]);
        
        if (is_wp_error($response)) {
            return false;
        }
        
        return json_decode(wp_remote_retrieve_body($response), true);
    }
    
    /**
     * Get compression result
     */
    public function get_result($job_id) {
        $response = wp_remote_get($this->api_url . '/result/' . $job_id, [
            'headers' => ['X-API-Key' => $this->api_key],
            'timeout' => 15
        ]);
        
        if (is_wp_error($response)) {
            return false;
        }
        
        return json_decode(wp_remote_retrieve_body($response), true);
    }
    
    /**
     * Check pending jobs (cron)
     */
    public function check_pending_jobs() {
        global $wpdb;
        
        $pending_attachments = $wpdb->get_results("
            SELECT post_id, meta_value as job_id 
            FROM {$wpdb->postmeta} 
            WHERE meta_key = '_compression_job_id'
            AND post_id IN (
                SELECT post_id FROM {$wpdb->postmeta}
                WHERE meta_key = '_compression_status'
                AND meta_value IN ('queued', 'processing')
            )
            LIMIT 50
        ");
        
        foreach ($pending_attachments as $item) {
            $status = $this->get_status($item->job_id);
            
            if ($status && $status['overall_status'] === 'completed') {
                $result = $this->get_result($item->job_id);
                
                if ($result && $result['overall_status'] === 'completed') {
                    update_post_meta($item->post_id, '_compression_status', 'completed');
                    update_post_meta($item->post_id, '_compression_result', $result);
                    
                    // Save compressed URL
                    if (isset($result['video_result']['compressed_url'])) {
                        update_post_meta($item->post_id, '_compressed_video_url', $result['video_result']['compressed_url']);
                    }
                    if (isset($result['image_result'])) {
                        update_post_meta($item->post_id, '_compressed_image_data', $result['image_result']);
                    }
                }
            } elseif ($status && $status['overall_status'] === 'failed') {
                update_post_meta($item->post_id, '_compression_status', 'failed');
            }
        }
    }
    
    /**
     * Add admin menu
     */
    public function add_admin_menu() {
        add_menu_page(
            'Video Compressor',
            'Compressor',
            'manage_options',
            'video-compressor',
            [$this, 'admin_page'],
            'dashicons-video-alt3',
            25
        );
        
        add_submenu_page(
            'video-compressor',
            'Settings',
            'Settings',
            'manage_options',
            'video-compressor-settings',
            [$this, 'settings_page']
        );
    }
    
    /**
     * Enqueue scripts
     */
    public function enqueue_scripts($hook) {
        if ($hook !== 'toplevel_page_video-compressor') {
            return;
        }
        
        wp_enqueue_script('jquery');
    }
    
    /**
     * Admin page
     */
    public function admin_page() {
        ?>
        <div class="wrap">
            <h1>Video Compression Status</h1>
            
            <table class="wp-list-table widefat fixed striped">
                <thead>
                    <tr>
                        <th>Media</th>
                        <th>Type</th>
                        <th>Status</th>
                        <th>Progress</th>
                        <th>Compressed URL</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <?php
                    $args = [
                        'post_type' => 'attachment',
                        'post_mime_type' => ['video', 'image'],
                        'posts_per_page' => 50,
                        'meta_query' => [
                            [
                                'key' => '_compression_job_id',
                                'compare' => 'EXISTS'
                            ]
                        ]
                    ];
                    
                    $media = get_posts($args);
                    
                    if (empty($media)) {
                        echo '<tr><td colspan="6">No compression jobs found. Upload a video or image to start!</td></tr>';
                    }
                    
                    foreach ($media as $item) {
                        $job_id = get_post_meta($item->ID, '_compression_job_id', true);
                        $status = get_post_meta($item->ID, '_compression_status', true);
                        $type = get_post_meta($item->ID, '_compression_type', true);
                        $compressed_url = get_post_meta($item->ID, '_compressed_video_url', true);
                        
                        echo '<tr>';
                        echo '<td>' . esc_html($item->post_title) . '</td>';
                        echo '<td>' . esc_html(ucfirst($type)) . '</td>';
                        echo '<td><span class="status-' . esc_attr($status) . '">' . esc_html(ucfirst($status)) . '</span></td>';
                        
                        if ($status === 'completed') {
                            echo '<td>100%</td>';
                            echo '<td>';
                            if ($compressed_url) {
                                echo '<a href="' . esc_url($compressed_url) . '" target="_blank">View Compressed</a>';
                            } else {
                                echo '-';
                            }
                            echo '</td>';
                        } else {
                            echo '<td>-</td>';
                            echo '<td>-</td>';
                        }
                        
                        echo '<td>';
                        echo '<button class="button check-status" data-job-id="' . esc_attr($job_id) . '" data-post-id="' . esc_attr($item->ID) . '">Check Status</button>';
                        echo '</td>';
                        echo '</tr>';
                    }
                    ?>
                </tbody>
            </table>
            
            <script>
            jQuery(document).ready(function($) {
                $('.check-status').on('click', function() {
                    var btn = $(this);
                    var jobId = btn.data('job-id');
                    var postId = btn.data('post-id');
                    
                    btn.prop('disabled', true).text('Checking...');
                    
                    $.post(ajaxurl, {
                        action: 'check_compression_status',
                        job_id: jobId,
                        post_id: postId
                    }, function(response) {
                        if (response.success) {
                            location.reload();
                        } else {
                            alert('Failed to check status');
                            btn.prop('disabled', false).text('Check Status');
                        }
                    });
                });
            });
            </script>
            
            <style>
                .status-completed { color: green; font-weight: bold; }
                .status-processing { color: orange; font-weight: bold; }
                .status-queued { color: blue; }
                .status-failed { color: red; font-weight: bold; }
            </style>
        </div>
        <?php
    }
    
    /**
     * Settings page
     */
    public function settings_page() {
        ?>
        <div class="wrap">
            <h1>Video Compressor Settings</h1>
            
            <form method="post" action="options.php">
                <?php settings_fields('video_compressor_options'); ?>
                
                <table class="form-table">
                    <tr>
                        <th scope="row">API URL</th>
                        <td>
                            <input type="text" name="vc_api_url" value="<?php echo esc_attr(get_option('vc_api_url', 'https://api.trendss.net/api')); ?>" class="regular-text" />
                        </td>
                    </tr>
                    <tr>
                        <th scope="row">API Key</th>
                        <td>
                            <input type="text" name="vc_api_key" value="<?php echo esc_attr(get_option('vc_api_key', '')); ?>" class="regular-text" />
                        </td>
                    </tr>
                    <tr>
                        <th scope="row">Auto Compress</th>
                        <td>
                            <label>
                                <input type="checkbox" name="vc_auto_compress" value="1" <?php checked(get_option('vc_auto_compress', true), true); ?> />
                                Automatically compress videos/images when uploaded
                            </label>
                        </td>
                    </tr>
                    <tr>
                        <th scope="row">Default Quality</th>
                        <td>
                            <select name="vc_default_quality">
                                <option value="low" <?php selected(get_option('vc_default_quality', 'medium'), 'low'); ?>>Low</option>
                                <option value="medium" <?php selected(get_option('vc_default_quality', 'medium'), 'medium'); ?>>Medium</option>
                                <option value="high" <?php selected(get_option('vc_default_quality', 'medium'), 'high'); ?>>High</option>
                                <option value="ultra" <?php selected(get_option('vc_default_quality', 'medium'), 'ultra'); ?>>Ultra</option>
                            </select>
                        </td>
                    </tr>
                </table>
                
                <?php submit_button(); ?>
            </form>
        </div>
        <?php
    }
    
    /**
     * AJAX: Compress media
     */
    public function ajax_compress_media() {
        $attachment_id = intval($_POST['attachment_id']);
        $type = sanitize_text_field($_POST['type']);
        
        $job_id = $this->compress_media($attachment_id, $type);
        
        if ($job_id) {
            wp_send_json_success(['job_id' => $job_id]);
        } else {
            wp_send_json_error('Failed to start compression');
        }
    }
    
    /**
     * AJAX: Check status
     */
    public function ajax_check_compression_status() {
        $job_id = sanitize_text_field($_POST['job_id']);
        $post_id = intval($_POST['post_id']);
        
        $status = $this->get_status($job_id);
        
        if ($status) {
            update_post_meta($post_id, '_compression_status', $status['overall_status']);
            
            if ($status['overall_status'] === 'completed') {
                $result = $this->get_result($job_id);
                if ($result && isset($result['video_result']['compressed_url'])) {
                    update_post_meta($post_id, '_compressed_video_url', $result['video_result']['compressed_url']);
                }
            }
            
            wp_send_json_success($status);
        } else {
            wp_send_json_error('Failed to check status');
        }
    }
    
    /**
     * AJAX: Get result
     */
    public function ajax_get_result() {
        $job_id = sanitize_text_field($_POST['job_id']);
        
        $result = $this->get_result($job_id);
        
        if ($result) {
            wp_send_json_success($result);
        } else {
            wp_send_json_error('Failed to get result');
        }
    }
}

// Initialize plugin
new VideoCompressorAPI();
